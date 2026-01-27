param(
  [switch] $DryRun
)

Set-StrictMode -Version Latest
$ErrorActionPreference = 'Stop'

$Here = Split-Path -Parent $MyInvocation.MyCommand.Path

function Normalize-Path([string] $p) {
  return ([IO.Path]::GetFullPath($p)).TrimEnd('\\')
}

$Target = Normalize-Path $Here
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

if ($Already) {
  Write-Output ("Already on User PATH: {0}" -f $Target)
  exit 0
}

$NewUserPath = if ($CurrentUserPath.Trim().Length -eq 0) { $Target } else { "$CurrentUserPath;$Target" }

if ($DryRun) {
  Write-Output 'Dry run (no changes made).'
  Write-Output ("Would set User PATH to include: {0}" -f $Target)
  exit 0
}

[Environment]::SetEnvironmentVariable('Path', $NewUserPath, 'User')
$env:Path = $NewUserPath

Write-Output ("Added to User PATH: {0}" -f $Target)
Write-Output 'Open a new terminal to pick up the change everywhere.'
