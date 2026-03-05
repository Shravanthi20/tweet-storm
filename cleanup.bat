@echo off
echo Killing any existing tweet-storm processes...
taskkill /f /im tweet-storm.exe 2>nul
echo Waiting for ports to be released...
timeout /t 2 /nobreak >nul
echo Done! Ports 8000-8004 are now free.
echo.
echo You can now start nodes:
echo   tweet-storm.exe -role=leader  -port=8000
echo   tweet-storm.exe -role=worker  -port=8001
echo   tweet-storm.exe -role=worker  -port=8002
echo   tweet-storm.exe -role=worker  -port=8003
echo   tweet-storm.exe -role=worker  -port=8004
echo   tweet-storm.exe -role=client
