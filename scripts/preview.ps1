param(
    [Parameter(Mandatory=$true)]
    [string]$PreviewDir,
    [Parameter(ValueFromRemainingArguments=$true)]
    [string[]]$Args
)

$env:CRUSH_GLOBAL_DATA = Join-Path $PreviewDir "data"
$starExe = Join-Path $PreviewDir "star.exe"

# Set title for clarity
$host.ui.RawUI.WindowTitle = "Star Preview"

# Execute
& $starExe $Args
