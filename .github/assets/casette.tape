# Used to generate demo gif for the README.md
# See https://github.com/charmbracelet/vhs
# Do not forget to mask sensitive information before running this script (E.g. MAC addresses, extra data etc.)

Set Padding 30
Set Shell zsh
Output demo.gif
Set FontSize 19
Set Width 1920
Set Height 1080
Set CursorBlink false
Set WindowBar Colorful
Set Margin 0
Set BorderRadius 10


Sleep 500ms
Type "whosthere --help"
Enter
Sleep 4.5s

Hide
Type "clear"
Enter
Sleep 500ms
Show

Type "whosthere"
Sleep 500ms
Enter
Sleep 1s
Type "jjj"
Sleep 500ms
Escape
Type "[27;5;116~"
Sleep 500ms
Type "k"
Enter
Sleep 1s
Type "jjjk"
Sleep 500ms
Type "/google"
Sleep 500ms
Enter
Sleep 1s
Type "jj"
Enter
Sleep 2s
Type "p"
Sleep 1.5s
Enter
Sleep 3.5s
Escape
Sleep 500ms
Escape
Sleep 1s

Hide
Escape
Type "[27;5;99~"
Sleep 500ms
Type "clear"
Sleep 500ms
Enter
Show

Type "whosthere daemon"
Sleep 500ms
Type " --port 8080"
Sleep 500ms
Enter
Sleep 6s
Ctrl+C

