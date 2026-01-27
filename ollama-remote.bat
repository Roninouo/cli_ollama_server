@echo off
setlocal EnableExtensions EnableDelayedExpansion

set "DEFAULT_HOST=http://10.65.117.238:11434"
set "HOST_OVERRIDE="
set "SHOW_HELP="
set "PASSTHRU="

:parse
if "%~1"=="" goto :after_parse

if /I "%~1"=="--help" (
  set "SHOW_HELP=1"
  shift
  goto :parse
)
if /I "%~1"=="-h" (
  set "SHOW_HELP=1"
  shift
  goto :parse
)

if /I "%~1"=="--host" (
  if "%~2"=="" (
    echo ERROR: --host requires a value. 1>&2
    exit /b 2
  )
  set "HOST_OVERRIDE=%~2"
  shift
  shift
  goto :parse
)

set "ARG=%~1"
if not "!ARG!"=="" (
  if /I "!ARG:~0,7!"=="--host=" (
    set "HOST_OVERRIDE=!ARG:~7!"
    shift
    goto :parse
  )
)

set "PASSTHRU=!PASSTHRU! ^"%~1^""
shift
goto :parse

:after_parse
if defined SHOW_HELP goto :help

if defined HOST_OVERRIDE (
  set "OLLAMA_HOST=%HOST_OVERRIDE%"
) else (
  if not defined OLLAMA_HOST set "OLLAMA_HOST=%DEFAULT_HOST%"
)

rem Avoid proxy issues for a local network endpoint (only when relevant)
if not defined NO_PROXY (
  echo %OLLAMA_HOST% | findstr /C:"10.65.117.238" >nul 2>nul
  if not errorlevel 1 set "NO_PROXY=10.65.117.238"
)

rem Prefer explicit OLLAMA_EXE if set, else use PATH
set "OLLAMA_CMD=ollama"
if defined OLLAMA_EXE set "OLLAMA_CMD=%OLLAMA_EXE%"

"%OLLAMA_CMD%" !PASSTHRU!
set "EXITCODE=%ERRORLEVEL%"
endlocal & exit /b %EXITCODE%

:help
echo Usage: ollama-remote [--host ^<url^>] ^<ollama-args...^>
echo.
echo Wraps the official Ollama CLI, defaulting OLLAMA_HOST to:
echo   %DEFAULT_HOST%
echo.
echo Environment overrides:
echo   OLLAMA_HOST  If set, the wrapper will not change it.
echo   OLLAMA_EXE   Full path to ollama.exe to use.
echo.
echo Examples:
echo   ollama-remote list
echo   ollama-remote run llama3:8b
echo   ollama-remote --host http://10.65.117.238:11434 ps
exit /b 0
