package teams

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/charmbracelet/crush/internal/lock"
	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskClaimed   TaskStatus = "claimed"
	TaskRunning   TaskStatus = "running"
	TaskBlocked   TaskStatus = "blocked"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
)

type AgentStatus string

const (
	AgentWorking AgentStatus = "working"
	AgentIdle    AgentStatus = "idle"
	AgentStopped AgentStatus = "stopped"
	AgentFailed  AgentStatus = "failed"
)

var (
	ErrNotFound          = errors.New("team object not found")
	ErrTaskClaimed       = errors.New("task is already claimed")
	ErrInvalidStatus     = errors.New("invalid task status")
	ErrInvalidTransition = errors.New("invalid task status transition")
)

type Agent struct {
	Name        string      `json:"name"`
	Role        string      `json:"role"`
	SessionID   string      `json:"session_id"`
	Status      AgentStatus `json:"status"`
	CurrentTask string      `json:"current_task,omitempty"`
	StartedAt   time.Time   `json:"started_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type Task struct {
	ID          string     `json:"id"`
	Subject     string     `json:"subject"`
	Description string     `json:"description,omitempty"`
	Status      TaskStatus `json:"status"`
	Owner       string     `json:"owner,omitempty"`
	BlockedBy   []string   `json:"blocked_by,omitempty"`
	Result      string     `json:"result,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Message struct {
	ID        string    `json:"id"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	ReadAt    time.Time `json:"read_at,omitempty"`
}

type State struct {
	Name      string    `json:"name"`
	Agents    []Agent   `json:"agents"`
	Tasks     []Task    `json:"tasks"`
	Messages  []Message `json:"messages"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Store struct{ dir string }

func NewStore(workingDir string) *Store {
	return &Store{dir: filepath.Join(workingDir, ".star", "tasks")}
}

func (s *Store) Join(ctx context.Context, teamName string, agent Agent) (State, error) {
	return s.update(ctx, teamName, func(state *State) error {
		now := time.Now().UTC()
		for i := range state.Agents {
			if state.Agents[i].Name == agent.Name {
				agent.StartedAt = state.Agents[i].StartedAt
				agent.UpdatedAt = now
				state.Agents[i] = agent
				return nil
			}
		}
		agent.StartedAt = now
		agent.UpdatedAt = now
		state.Agents = append(state.Agents, agent)
		return nil
	})
}

func (s *Store) Leave(ctx context.Context, teamName, agentName string, status AgentStatus) (State, error) {
	return s.update(ctx, teamName, func(state *State) error {
		for i := range state.Agents {
			if state.Agents[i].Name == agentName {
				state.Agents[i].Status = status
				state.Agents[i].CurrentTask = ""
				state.Agents[i].UpdatedAt = time.Now().UTC()
				return nil
			}
		}
		return ErrNotFound
	})
}

func (s *Store) ListAgents(ctx context.Context, teamName string) ([]Agent, error) {
	state, err := s.read(ctx, teamName)
	if err != nil {
		return nil, err
	}
	return state.Agents, nil
}

func (s *Store) CreateTask(ctx context.Context, teamName, subject, description string, blockedBy []string) (Task, error) {
	var task Task
	_, err := s.update(ctx, teamName, func(state *State) error {
		now := time.Now().UTC()
		task = Task{ID: uuid.NewString(), Subject: strings.TrimSpace(subject), Description: strings.TrimSpace(description), Status: TaskPending, BlockedBy: slices.Clone(blockedBy), CreatedAt: now, UpdatedAt: now}
		state.Tasks = append(state.Tasks, task)
		return nil
	})
	return task, err
}

func (s *Store) ListTasks(ctx context.Context, teamName string) ([]Task, error) {
	state, err := s.read(ctx, teamName)
	if err != nil {
		return nil, err
	}
	return state.Tasks, nil
}

func (s *Store) ClaimTask(ctx context.Context, teamName, taskID, agentName string) (Task, error) {
	var claimed Task
	_, err := s.update(ctx, teamName, func(state *State) error {
		for i := range state.Tasks {
			if state.Tasks[i].ID != taskID {
				continue
			}
			if state.Tasks[i].Owner != "" && state.Tasks[i].Owner != agentName {
				return ErrTaskClaimed
			}
			if state.Tasks[i].Status != TaskPending && state.Tasks[i].Status != TaskClaimed {
				return fmt.Errorf("%w: cannot claim %s task", ErrInvalidTransition, state.Tasks[i].Status)
			}
			if blocked(state.Tasks[i], state.Tasks) {
				return fmt.Errorf("%w: task dependencies are incomplete", ErrInvalidTransition)
			}
			state.Tasks[i].Owner = agentName
			state.Tasks[i].Status = TaskClaimed
			state.Tasks[i].UpdatedAt = time.Now().UTC()
			claimed = state.Tasks[i]
			setAgentTask(state, agentName, taskID, AgentWorking)
			return nil
		}
		return ErrNotFound
	})
	return claimed, err
}

func (s *Store) UpdateTask(ctx context.Context, teamName, taskID, agentName string, status TaskStatus, result string) (Task, error) {
	if !validTaskStatus(status) {
		return Task{}, ErrInvalidStatus
	}
	var updated Task
	_, err := s.update(ctx, teamName, func(state *State) error {
		for i := range state.Tasks {
			if state.Tasks[i].ID != taskID {
				continue
			}
			if state.Tasks[i].Owner != "" && state.Tasks[i].Owner != agentName {
				return ErrTaskClaimed
			}
			if !canTransition(state.Tasks[i].Status, status) {
				return fmt.Errorf("%w: %s to %s", ErrInvalidTransition, state.Tasks[i].Status, status)
			}
			state.Tasks[i].Owner = agentName
			state.Tasks[i].Status = status
			state.Tasks[i].Result = result
			state.Tasks[i].UpdatedAt = time.Now().UTC()
			updated = state.Tasks[i]
			agentStatus, currentTask := AgentWorking, taskID
			if status == TaskCompleted || status == TaskFailed || status == TaskBlocked {
				agentStatus, currentTask = AgentIdle, ""
			}
			setAgentTask(state, agentName, currentTask, agentStatus)
			return nil
		}
		return ErrNotFound
	})
	return updated, err
}

func (s *Store) SendMessage(ctx context.Context, teamName string, message Message) (Message, error) {
	_, err := s.update(ctx, teamName, func(state *State) error {
		if message.To != "all" && !hasAgent(state.Agents, message.To) {
			return fmt.Errorf("%w: agent %q", ErrNotFound, message.To)
		}
		message.ID = uuid.NewString()
		message.CreatedAt = time.Now().UTC()
		state.Messages = append(state.Messages, message)
		return nil
	})
	return message, err
}

func (s *Store) Inbox(ctx context.Context, teamName, agentName string, unreadOnly bool, markRead bool) ([]Message, error) {
	var messages []Message
	_, err := s.update(ctx, teamName, func(state *State) error {
		now := time.Now().UTC()
		for i := range state.Messages {
			message := &state.Messages[i]
			if message.To != agentName && message.To != "all" {
				continue
			}
			if unreadOnly && !message.ReadAt.IsZero() {
				continue
			}
			messages = append(messages, *message)
			if markRead && message.ReadAt.IsZero() {
				message.ReadAt = now
			}
		}
		return nil
	})
	return messages, err
}

func (s *Store) read(ctx context.Context, teamName string) (State, error) {
	var state State
	_, err := s.update(ctx, teamName, func(current *State) error {
		state = *current
		state.Agents = slices.Clone(current.Agents)
		state.Tasks = slices.Clone(current.Tasks)
		state.Messages = slices.Clone(current.Messages)
		return nil
	})
	return state, err
}

func (s *Store) update(ctx context.Context, teamName string, fn func(*State) error) (State, error) {
	teamName, err := safeName(teamName)
	if err != nil {
		return State{}, err
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return State{}, fmt.Errorf("create team directory: %w", err)
	}
	path := filepath.Join(s.dir, teamName+".json")
	release, err := lock.File(ctx, path+".lock")
	if err != nil {
		return State{}, fmt.Errorf("lock team: %w", err)
	}
	defer release()
	state := State{Name: teamName, CreatedAt: time.Now().UTC()}
	data, err := os.ReadFile(path)
	if err == nil {
		if err := json.Unmarshal(data, &state); err != nil {
			return State{}, fmt.Errorf("decode team: %w", err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return State{}, fmt.Errorf("read team: %w", err)
	}
	if err := fn(&state); err != nil {
		return State{}, err
	}
	state.UpdatedAt = time.Now().UTC()
	data, err = json.MarshalIndent(state, "", "  ")
	if err != nil {
		return State{}, fmt.Errorf("encode team: %w", err)
	}
	tmp, err := os.CreateTemp(s.dir, teamName+"-*.tmp")
	if err != nil {
		return State{}, fmt.Errorf("create team temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return State{}, fmt.Errorf("set team file permissions: %w", err)
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return State{}, fmt.Errorf("write team: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return State{}, fmt.Errorf("close team: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return State{}, fmt.Errorf("replace team: %w", err)
	}
	return state, nil
}

func safeName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("team name is required")
	}
	for _, char := range name {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '-' || char == '_' {
			continue
		}
		return "", errors.New("team name may only contain letters, numbers, dashes, and underscores")
	}
	return name, nil
}

func hasAgent(agents []Agent, name string) bool {
	return slices.ContainsFunc(agents, func(agent Agent) bool { return agent.Name == name })
}

func setAgentTask(state *State, agentName, taskID string, status AgentStatus) {
	for i := range state.Agents {
		if state.Agents[i].Name == agentName {
			state.Agents[i].CurrentTask = taskID
			state.Agents[i].Status = status
			state.Agents[i].UpdatedAt = time.Now().UTC()
			return
		}
	}
}

func blocked(task Task, tasks []Task) bool {
	for _, dependency := range task.BlockedBy {
		index := slices.IndexFunc(tasks, func(candidate Task) bool { return candidate.ID == dependency })
		if index < 0 || tasks[index].Status != TaskCompleted {
			return true
		}
	}
	return false
}

func validTaskStatus(status TaskStatus) bool {
	return slices.Contains([]TaskStatus{TaskPending, TaskClaimed, TaskRunning, TaskBlocked, TaskCompleted, TaskFailed}, status)
}

func canTransition(from, to TaskStatus) bool {
	if from == to {
		return true
	}
	switch from {
	case TaskPending:
		return to == TaskClaimed || to == TaskRunning || to == TaskBlocked || to == TaskFailed
	case TaskClaimed:
		return to == TaskPending || to == TaskRunning || to == TaskBlocked || to == TaskFailed
	case TaskRunning:
		return to == TaskBlocked || to == TaskCompleted || to == TaskFailed
	case TaskBlocked:
		return to == TaskPending || to == TaskRunning || to == TaskFailed
	case TaskCompleted, TaskFailed:
		return false
	default:
		return false
	}
}
