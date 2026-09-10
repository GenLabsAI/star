Coordinate with other agents through shared teams. Teams persist across sessions and support task management and messaging.

## Actions

### join
Register this agent as a member of a team. Each agent has a name and a role (e.g., architect, coder, reviewer, researcher, tester). Re-joining updates the agent's status.

### leave
Unregister this agent from the team and set its status to stopped.

### list_agents
List all agents currently registered in a team, with their roles, statuses, and current tasks.

### create_task
Create a new task in the team's shared queue. Tasks have a subject, optional description, and optional `blocked_by` list of task IDs that must complete first.

### list_tasks
List all tasks in the team, including their status, owner, and results.

### claim_task
Claim an available (pending) task. The task must not be blocked by incomplete dependencies. Only one agent can own a task at a time.

### update_task
Update a task's status. Valid statuses: `pending`, `claimed`, `running`, `blocked`, `completed`, `failed`. Include a `result` when completing or failing a task.

Task lifecycle: pending -> claimed -> running -> completed/failed. Tasks can also be marked blocked and later unblocked.

### send_message
Send a message to another agent by name, or to `all` for broadcast. Messages are stored persistently and delivered via the inbox.

### inbox
Read messages sent to this agent. Use `unread_only: true` to see only new messages. Use `mark_read: true` to mark them as read.

## Parameters

- `action` (required): One of the actions above.
- `team` (required): Team name (letters, numbers, dashes, underscores only).
- `agent`: Agent name. Defaults to the session ID if omitted.
- `role`: Agent role for join (e.g., architect, coder, reviewer). Defaults to "coder".
- `subject`: Task subject (required for create_task).
- `description`: Task description (optional for create_task).
- `task_id`: Task ID (required for claim_task and update_task).
- `status`: New task status (required for update_task).
- `result`: Task result text (optional for update_task).
- `blocked_by`: List of task IDs this task depends on (optional for create_task).
- `to`: Recipient agent name or "all" (required for send_message).
- `message`: Message content (required for send_message).
- `unread_only`: Only return unread messages (optional for inbox).
- `mark_read`: Mark returned messages as read (optional for inbox).

## Typical workflow

1. Each agent joins the team with a role.
2. One agent (often the architect) creates tasks.
3. Agents claim and work on tasks, updating status as they go.
4. Agents send messages to coordinate, ask questions, or share results.
5. Agents check their inbox for messages from teammates.
