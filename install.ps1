param(
  [switch] $DryRun
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Here = Split-Path -Parent $MyInvocation.MyCommand.Path

function Normalize-Path([string] $p) {
  return ([IO.Path]::GetFullPath($p)).TrimEnd('\\')
}

function Get-InstallTarget([string] $here) {
  $candidates = @(
    $here,
    (Join-Path $here 'dist'),
    (Join-Path $here 'dist\\windows-amd64')
  )

  foreach ($d in $candidates) {
    try {
      $exe = Join-Path $d 'ollama-remote.exe'
      if (Test-Path -LiteralPath $exe) {
        return $d
      }
    } catch {
      # ignore
    }
  }

  return $here
}

$Target = Normalize-Path (Get-InstallTarget $Here)
$CurrentUserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($null -eq $CurrentUserPath) { $CurrentUserPath = '' }

$Parts = @()
if ($CurrentUserPath.Trim().Length -gt 0) {
  $Parts = $CurrentUserPath.Split(';') | Where-Object { $_ -and $_.Trim().Length -gt 0 }
}

$Already = $false
foreach ($p in $Parts) {
  try {
    if ((Normalize-Path $p) -ieq $Target) { $Already = $true; break }
  } catch {
    # ignore invalid path segments
  }
}

$CurrentProcPath = $env:Path
if ($null -eq $CurrentProcPath) { $CurrentProcPath = '' }
$ProcParts = @()
if ($CurrentProcPath.Trim().Length -gt 0) {
  $ProcParts = $CurrentProcPath.Split(';') | Where-Object { $_ -and $_.Trim().Length -gt 0 }
}

$AlreadyProc = $false
foreach ($p in $ProcParts) {
  try {
    if ((Normalize-Path $p) -ieq $Target) { $AlreadyProc = $true; break }
  } catch {
    # ignore invalid path segments
  }
}

if ($Already -and $AlreadyProc) {
  Write-Output ("Already on PATH: {0}" -f $Target)
  exit 0
}

if ($Already -and (-not $AlreadyProc)) {
  if ($DryRun) {
    Write-Output 'Dry run (no changes made).'
    Write-Output ("Already on User PATH; would update current session PATH to include: {0}" -f $Target)
    exit 0
  }

  $env:Path = if ($CurrentProcPath.Trim().Length -eq 0) { $Target } else { "$CurrentProcPath;$Target" }
  Write-Output ("Already on User PATH; updated current session PATH: {0}" -f $Target)
  Write-Output 'Open a new terminal to pick up the change everywhere.'
  exit 0
}

$NewUserPath = if ($CurrentUserPath.Trim().Length -eq 0) { $Target } else { "$CurrentUserPath;$Target" }

if ($DryRun) {
  Write-Output 'Dry run (no changes made).'
  Write-Output ("Would set User PATH to include: {0}" -f $Target)
  exit 0
}

[Environment]::SetEnvironmentVariable('Path', $NewUserPath, 'User')
$env:Path = if ($CurrentProcPath.Trim().Length -eq 0) { $Target } else { "$CurrentProcPath;$Target" }

Write-Output ("Added to User PATH: {0}" -f $Target)
Write-Output 'Open a new terminal to pick up the change everywhere.'
