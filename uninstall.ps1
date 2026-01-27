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

$Targets = New-Object System.Collections.Generic.List[string]
$Targets.Add((Normalize-Path $Here))
try { $Targets.Add((Normalize-Path (Get-InstallTarget $Here))) } catch { }

# de-dupe
$UniqueTargets = New-Object System.Collections.Generic.List[string]
foreach ($t in $Targets) {
  $seen = $false
  foreach ($u in $UniqueTargets) {
    if ($u -ieq $t) { $seen = $true; break }
  }
  if (-not $seen) { $UniqueTargets.Add($t) }
}

$TargetDisplay = ($UniqueTargets -join ', ')
$CurrentUserPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($null -eq $CurrentUserPath) { $CurrentUserPath = '' }

$Parts = @()
if ($CurrentUserPath.Trim().Length -gt 0) {
  $Parts = $CurrentUserPath.Split(';') | Where-Object { $_ -and $_.Trim().Length -gt 0 }
}

$Kept = New-Object System.Collections.Generic.List[string]
foreach ($p in $Parts) {
  $keep = $true
  try {
    foreach ($t in $UniqueTargets) {
      if ((Normalize-Path $p) -ieq $t) { $keep = $false; break }
    }
  } catch {
    # keep invalid path segments as-is
  }
  if ($keep) { $Kept.Add($p) }
}

$NewUserPath = ($Kept -join ';')

if ($DryRun) {
  Write-Output 'Dry run (no changes made).'
  Write-Output ("Would remove from User PATH: {0}" -f $TargetDisplay)
  exit 0
}

[Environment]::SetEnvironmentVariable('Path', $NewUserPath, 'User')

$CurrentProcPath = $env:Path
if ($null -eq $CurrentProcPath) { $CurrentProcPath = '' }
$ProcParts = @()
if ($CurrentProcPath.Trim().Length -gt 0) {
  $ProcParts = $CurrentProcPath.Split(';') | Where-Object { $_ -and $_.Trim().Length -gt 0 }
}

$ProcKept = New-Object System.Collections.Generic.List[string]
foreach ($p in $ProcParts) {
  $keep = $true
  try {
    foreach ($t in $UniqueTargets) {
      if ((Normalize-Path $p) -ieq $t) { $keep = $false; break }
    }
  } catch {
    # keep invalid path segments as-is
  }
  if ($keep) { $ProcKept.Add($p) }
}

$env:Path = ($ProcKept -join ';')

Write-Output ("Removed from User PATH: {0}" -f $TargetDisplay)
Write-Output 'Open a new terminal to pick up the change everywhere.'
