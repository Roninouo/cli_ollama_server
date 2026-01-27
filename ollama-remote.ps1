param(
  [Parameter(ValueFromRemainingArguments = $true)]
  [string[]] $OllamaArgs
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

if ($null -eq $OllamaArgs) { $OllamaArgs = @() }

$DefaultHost = 'http://10.65.117.238:11434'

function Show-Help {
  Write-Output 'Usage: ollama-remote.ps1 [--host <url>] <ollama-args...>'
  Write-Output ''
  Write-Output 'Wraps the official Ollama CLI, defaulting OLLAMA_HOST to:'
  Write-Output ("  {0}" -f $DefaultHost)
  Write-Output ''
  Write-Output 'Environment overrides:'
  Write-Output '  OLLAMA_HOST  If set, the wrapper will not change it.'
  Write-Output '  OLLAMA_EXE   Full path to ollama.exe to use.'
  Write-Output ''
  Write-Output 'Examples:'
  Write-Output '  .\ollama-remote.ps1 list'
  Write-Output '  .\ollama-remote.ps1 run llama3:8b'
  Write-Output '  .\ollama-remote.ps1 --host http://10.65.117.238:11434 ps'
}

$HostOverride = $null
$Filtered = New-Object System.Collections.Generic.List[string]

for ($i = 0; $i -lt $OllamaArgs.Count; $i++) {
  $a = $OllamaArgs[$i]

  if ($a -eq '--help' -or $a -eq '-h') {
    Show-Help
    exit 0
  }

  if ($a -eq '--host') {
    if (($i + 1) -ge $OllamaArgs.Count) {
      Write-Error 'ERROR: --host requires a value.'
      exit 2
    }
    $HostOverride = $OllamaArgs[$i + 1]
    $i++
    continue
  }

  if ($a -like '--host=*') {
    $HostOverride = $a.Substring(7)
    continue
  }

  $Filtered.Add($a)
}

if ([string]::IsNullOrWhiteSpace($HostOverride)) {
  if ([string]::IsNullOrWhiteSpace($env:OLLAMA_HOST)) {
    $env:OLLAMA_HOST = $DefaultHost
  }
} else {
  $env:OLLAMA_HOST = $HostOverride
}

if ([string]::IsNullOrWhiteSpace($env:NO_PROXY)) {
  $env:NO_PROXY = '10.65.117.238'
}

$OllamaExe = $env:OLLAMA_EXE
if ([string]::IsNullOrWhiteSpace($OllamaExe)) {
  $cmd = Get-Command ollama -ErrorAction SilentlyContinue
  if ($cmd) {
    $OllamaExe = $cmd.Source
  }
}

if ([string]::IsNullOrWhiteSpace($OllamaExe) -or -not (Test-Path -LiteralPath $OllamaExe)) {
  Write-Error 'ollama.exe not found. Install Ollama or set OLLAMA_EXE to the full path.'
  exit 127
}

& $OllamaExe @Filtered
exit $LASTEXITCODE
