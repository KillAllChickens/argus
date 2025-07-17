@echo off
setlocal

set "APP_NAME=argus"
set "CONFIG_DIR=%APPDATA%\%APP_NAME%"

echo [*] Building %APP_NAME% for Windows...

:: Setting up environment for Go build
set "GOOS=windows"
if /I "%PROCESSOR_ARCHITECTURE%" == "AMD64" (
    set "GOARCH=amd64"
) else if /I "%PROCESSOR_ARCHITECTURE%" == "ARM64" (
    set "GOARCH=arm64"
) else (
    echo [!] Unsupported architecture: %PROCESSOR_ARCHITECTURE%
    goto :eof
)

echo [*] Target: %GOOS%/%GOARCH%

:: Building the Go executable
go install .
if %errorlevel% neq 0 (
    echo [!] Build failed. Check the errors above.
    goto :eof
)
echo [✔] Build successful.

echo.
echo [*] The installation directory will be: %INSTALL_DIR%
echo [*] The configuration directory will be: %CONFIG_DIR%
echo.

echo [*] Ensuring config directory exists...
mkdir "%CONFIG_DIR%" 2>nul

echo [*] Copying configuration files...
if exist ".\config" (
    xcopy /E /I /Y ".\config" "%CONFIG_DIR%\"
) else (
    echo [!] No 'config' directory found to copy. Skipping.
)

set "GOBIN=%USERPROFILE%\go\bin"
echo.
echo [*] Your binary is located at: %GOBIN%\%APP_NAME%.exe

:: Check if GOBIN is in PATH
echo %PATH% | find /I "%GOBIN%" >nul
if %errorlevel% neq 0 (
    echo [!] %GOBIN% is not in your PATH.
    echo     Please add it manually or re-run the Go installer with that option.
) else (
    echo [✔] Your Go bin path is already in PATH.
)

echo.
echo [✔] Installation complete!
echo.
echo You can now run '%APP_NAME% config' or '%APP_NAME% c'.
echo (You might need to open a new terminal for the PATH change to take effect).
echo.

endlocal
