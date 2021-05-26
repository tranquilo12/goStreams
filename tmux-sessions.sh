#!/bin/sh
tmux new-session \; \
send-keys 'fish' C-m\; \
send-keys 'clear' C-m\; \
split-window -v \; \
send-keys 'fish' C-m \; \
send-keys 'clear' C-m\; \
split-window -h \; \
send-keys 'fish' C-m \; \
send-keys 'clear' C-m\; \
select-pane -t 0 \; \
split-window -h \; \
send-keys 'fish' C-m \; \
send-keys 'clear' C-m\; \
resize-pane -y 20 \; \
split-window -v \; \
resize-pane -y 50 \; \
send-keys 'fish' C-m \; \
send-keys 'clear' C-m\; \
