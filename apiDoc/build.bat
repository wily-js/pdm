@echo off
CD %~dp0
set distdir=%cd%\target
(apidoc -i ../controller -o target) & (start "C:\Program Files\Google\Chrome\Application\chrome.exe" %distdir%\index.html)
