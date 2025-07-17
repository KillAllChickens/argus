@echo off
setlocal

set "APP_NAME=argus"
set "INSTALL_DIR=%LOCALAPPDATA%\Programs\%APP_NAME%"
set "CONFIG_DIR=%APPDATA%\%APP_NAME%"

echo [*] Building %APP_NAME% for Windows...

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

go build -o "%APP_NAME%.exe" .
if %errorlevel% neq 0 (
    echo [!] Build failed. Check the errors above.
    goto :eof
)
echo [✔] Build successful.

echo [*] Ensuring installation directory exists...
mkdir "%INSTALL_DIR%" 2>nul

echo [*] Ensuring config directory exists...
mkdir "%CONFIG_DIR%" 2>nul

echo [*] Installing %APP_NAME%.exe...
move /Y "%APP_NAME%.exe" "%INSTALL_DIR%\"
if %errorlevel% neq 0 (
    echo [!] Failed to move the executable.
    goto :eof
)

echo [*] Copying configuration files...
if exist ".\config" (
    xcopy /E /I /Y ".\config" "%CONFIG_DIR%\"
) else (
    echo [!] No 'config' directory found to copy. Skipping.
)

echo [*] Checking if %INSTALL_DIR% is in your PATH...
echo %PATH% | find /I "%INSTALL_DIR%" >nul
if %errorlevel% neq 0 (
    echo.
    echo [!] WARNING: %INSTALL_DIR% is not in your PATH.
    echo     To run '%APP_NAME%' from anywhere, you need to add it.
    echo.
    set /p "do_setx=Add to PATH now? (y/n): "
    if /I "%do_setx%" == "y" (
        echo [*] Adding to PATH using setx...
        setx PATH "%INSTALL_DIR%;%PATH%"
        echo [✔] PATH updated. Please open a new Command Prompt or PowerShell.
    ) else (
        echo [*] OK. You can add it manually later.
    )
) else (
    echo.
    echo [✔] Already in PATH. Good to go.
)

echo.
echo [✔] All done! You can now run '%APP_NAME% configure' or '%APP_NAME% c'.
echo     (You might need to open a new terminal window first).
echo.

endlocal
