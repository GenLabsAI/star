package logo

import (
	"testing"

	"github.com/charmbracelet/crush/internal/ui/styles"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/require"
)

func TestRender(t *testing.T) {
	// Use the default styles to test the rendering function
	s := styles.CharmtonePantera()
	o := Opts{
		FieldColor:   s.Logo.FieldColor,
		TitleColorA:  s.Logo.TitleColorA,
		TitleColorB:  s.Logo.TitleColorB,
		CharmColor:   s.Logo.CharmColor,
		VersionColor: s.Logo.VersionColor,
		Width:        120,
	}

	// Normal render (not hyper, wide)
	logo := ansi.Strip(Render(s.Logo.GradCanvas, "v1.2.3", false, o))
	require.Contains(t, logo, "v1.2.3")
	require.Contains(t, logo, "╭──╮╶─┬─╴╭──╮ ╭──╮")
	require.NotContains(t, logo, "█▀▀▀▄") // ensure Hyper logo is absent

	// Hyper render (wide)
	o.Hyper = true
	logoHyper := ansi.Strip(Render(s.Logo.GradCanvas, "v1.2.3", false, o))
	require.Contains(t, logoHyper, "╭──╮╶─┬─╴╭──╮ ╭──╮")
	require.Contains(t, logoHyper, "█▀▀▀▄") // Hyper (H, Y, P, E, R) should be present

	// Hyper render (compact)
	logoHyperCompact := ansi.Strip(Render(s.Logo.GradCanvas, "v1.2.3", true, o))
	require.Contains(t, logoHyperCompact, "╭──╮╶─┬─╴╭──╮ ╭──╮")
	require.Contains(t, logoHyperCompact, "█▀▀▀▄")
}

func TestSmallRender(t *testing.T) {
	s := styles.CharmtonePantera()
	o := Opts{
		Hyper: false,
	}
	small := ansi.Strip(SmallRender(&s, 20, o))
	require.Contains(t, small, "Star")
	require.NotContains(t, small, "HYPERSTAR")

	o.Hyper = true
	smallHyper := ansi.Strip(SmallRender(&s, 30, o))
	require.Contains(t, smallHyper, "HYPERSTAR")
}
