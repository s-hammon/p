#
# ~/.bashrc
#

# If not running interactively, don't do anything
[[ $- != *i* ]] && return

alias ls='ls --color=auto'
alias grep='grep --color=auto'
PS1='[\u@\h \W]\$ '

export LS_COLORS="*.go=00;36:*.py=00;33:*.sql=1;36:"

# Generated for envman. Do not edit.
[ -s "$HOME/.config/envman/load.sh" ] && source "$HOME/.config/envman/load.sh"

export PATH=$PATH:/usr/local/go/bin
export PATH=$PATH:$HOME/go/bin

# Turso
export PATH="$PATH:/home/sven/.turso"
export EDITOR=nvim

. "$HOME/.cargo/env"
export PATH=~/bin:$PATH

# The next line updates PATH for the Google Cloud SDK.
if [ -f '/home/sven/google-cloud-sdk/path.bash.inc' ]; then . '/home/sven/google-cloud-sdk/path.bash.inc'; fi

# The next line enables shell command completion for gcloud.
if [ -f '/home/sven/google-cloud-sdk/completion.bash.inc' ]; then . '/home/sven/google-cloud-sdk/completion.bash.inc'; fi

export PYENV_ROOT="$HOME/.pyenv"
export PATH="$PYENV_ROOT/bin:$PATH"
eval "$(pyenv init --path)"
eval "$(pyenv init -)"

# Pyenv stuff
pyenv_activate() {
    bname=${PWD##*/}
    pyenv activate "$bname"
    echo "venv: "$bname""
}
alias pyenv-activate=pyenv_activate

pyenv_make() {
    local python_version

    # look for file version
    if [[ -f .python_version ]]; then
        python_version="$(head -n 1 .python-version)"
    # look for local version
    elif
        python_version="$(pyenv local 2>/dev/null | head -n 1)"
        [[ -n "$python-version" ]]
    then
        :
    # default to global
    elif
        python_version="$(pyenv global 2>/dev/null | head -n 1)"
        [[ -n "$python-version" ]]
    then
        :
    else
        echo "Error: Could not determine Python version (.python-version, pyenv local, or pyenv global not found)." >&2
        return 1
    fi

    echo "Using Python $python_version to create virtualenv: ${PWD##*/}"
    pyenv virtualenv "$python_version" "${PWD##*/}" || return 1
    pyenv activate "${PWD##*/}" || return 1
}

alias pyenv-make=pyenv_make
