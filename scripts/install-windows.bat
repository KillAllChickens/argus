@echo off
setlocal

set "APP_NAME=argus"
set "INSTALL_DIR=%LOCALAPPDATA%\Programs\%APP_NAME%"
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
go build -o "%APP_NAME%.exe" .
if %errorlevel% neq 0 (
    echo [!] Build failed. Check the errors above.
    goto :eof
)
echo [✔] Build successful.

echo.
echo [*] The installation directory will be: %INSTALL_DIR%
echo [*] The configuration directory will be: %CONFIG_DIR%
echo.

echo [*] Ensuring installation directory exists...
mkdir "%INSTALL_DIR%" 2>nul

echo [*] Ensuring config directory exists...
mkdir "%CONFIG_DIR%" 2>nul

echo [*] Installing %APP_NAME%.exe...
move /Y "%APP_NAME%.exe" "%INSTALL_DIR%\"
if %errorlevel% neq 0 (
    echo [!] Failed to move the executable to %INSTALL_DIR%
    goto :eof
)

echo [*] Copying configuration files...
if exist ".\config" (
    xcopy /E /I /Y ".\config" "%CONFIG_DIR%\"
) else (
    echo [!] No 'config' directory found to copy. Skipping.
)

echo [*] Checking if the installation directory is in your user PATH...

set "KEY_NAME=HKEY_CURRENT_USER\Environment"
set "VALUE_NAME=Path"
set "NEEDS_PATH_UPDATE=true"

for /f "tokens=2,*" %%a in ('reg query "%KEY_NAME%" /v "%VALUE_NAME%" 2^>nul') do (
    echo "%%b" | find /I "%INSTALL_DIR%" >nul
    if %errorlevel% equ 0 (
        set "NEEDS_PATH_UPDATE="
    )
)

if defined NEEDS_PATH_UPDATE (
    echo.
    echo [!] WARNING: %INSTALL_DIR% is not in your user PATH.
    echo     To run '%APP_NAME%' from anywhere, you need to add it.
    echo.
    set /p "do_setx=Would you like to add it to your user PATH now? (y/n): "
    if /I "%do_setx%" == "y" (
        echo [*] Adding to user PATH. This will affect new terminals.

        for /f "tokens=2,*" %%a in ('reg query "%KEY_NAME%" /v "%VALUE_NAME%" 2^>nul') do (
            set "CURRENT_USER_PATH=%%b"
        )

        if defined CURRENT_USER_PATH (
            reg add "%KEY_NAME%" /v "%VALUE_NAME%" /t REG_EXPAND_SZ /d "%INSTALL_DIR%;%CURRENT_USER_PATH%" /f >nul
        ) else (
            reg add "%KEY_NAME%" /v "%VALUE_NAME%" /t REG_EXPAND_SZ /d "%INSTALL_DIR%" /f >nul
        )

        echo [✔] PATH updated. Please open a new Command Prompt or PowerShell to use the command.
    ) else (
        echo [*] OK. You can add it manually later if you wish.
    )
) else (
    echo.
    echo [✔] Installation directory is already in your user PATH.
)

echo.
echo [✔] Installation complete!
echo.
echo You can now run '%APP_NAME% configure' or '%APP_NAME% c'.
echo (You might need to open a new terminal for the PATH change to take effect).
echo.

:: Ask user if they want to open the installation directory
set /p "open_folder=Open the installation folder in File Explorer? (y/n): "
if /I "%open_folder%" == "y" (
    echo [*] Opening %INSTALL_DIR%...
    explorer "%INSTALL_DIR%"
)

endlocal
