FROM alpine
ENV HOME /app
WORKDIR $HOME
COPY todo-back $HOME/todo-back
EXPOSE 4000
ENTRYPOINT GIN_MODE=release $HOME/todo-back
