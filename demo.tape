# VHS demo for gh-list-pr
# Run: vhs demo.tape

Output demo.gif

Set FontSize 16
Set Width 1400
Set Height 700
Set Padding 10
Set TypingSpeed 50ms

Type "cd ~/git/github.com/neovim/neovim"
Enter
Sleep 500ms

Type "gh list-pr"
Enter
Sleep 8s

Type "feat"
Sleep 2s

Backspace 4
Sleep 500ms

Type "fix"
Sleep 2s

Ctrl+C
Sleep 1s
