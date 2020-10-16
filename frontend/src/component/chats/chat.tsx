import React, { useRef, useEffect, useReducer } from 'react';
import { useParams } from 'react-router-dom'
import { makeStyles, Grid, List, TextField, IconButton, ListItem, ListItemAvatar, ListItemText, Avatar, Typography } from '@material-ui/core'
import SendIcon from '@material-ui/icons/Send';
import Sidebar from './sidebar'

import { hostApi } from '../../utils/config';
import { User, ServerLog, Message } from '../../type'
import { ActionType, setBuffer, setConn, setMessage } from '../../actions/action'

type ChatsState = {
  buffer: string,
  conn: WebSocket | null,
  msg: Message[],
  users: User[],
}

type Action = {
  type: ActionType,
  payload: any,
}

const initialState = {
  buffer: "",
  conn: null,
  msg: [],
  users: [],
}

const reducer: React.Reducer<ChatsState, Action> = (state: ChatsState, action: Action) => {
  switch (action.type) {
    case 'BUFFER':
      return { ...state, buffer: action.payload }
    case 'CONNECTION':
      return { ...state, conn: action.payload }
    case 'RECIVED_LOG':
      switch (action.payload.command) {
        case 'NEW_USER':
          const { msg } = action.payload
          return { ...state, users: msg }
        case 'MESSAGE':
          state.msg.push(action.payload.msg)
          return { ...state, msg: state.msg }
        case 'CHANNEL_LOG':
          return { ...state, msg: action.payload.msg }
        default:
          return state
      }
    default:
      throw new Error();
  }
}

function createNotification(log: string) {
  return {
    command: "MESSAGE",
    msg: {
      username: "From Web Server",
      text: log,
    }
  }
}

const Chat: React.FC = () => {
  const classes = useStyles()
  let { channel } = useParams() as { channel: string }
  const [state, dispatch] = useReducer(reducer, initialState)
  const scrollRef = useRef<HTMLUListElement>(null)

  const sendMsg: VoidFunction = () => {
    const msg = state.buffer.trim().replace(/\r?\n/g, '<br>')
    if (msg === "") { dispatch(setBuffer("")); return }
    dispatch(setBuffer(""))

    const data = {
      msg: msg,
      channel: channel,
    }
    state.conn ? state.conn.send(JSON.stringify(data)) : dispatch(setMessage(createNotification("Connection is not exist.")))
  }

  const enterSendMsg = (e: any) => {
    if (e.keyCode === 13 && (e.metaKey || e.ctrlKey)) {
      e.preventDefault()
      sendMsg()
    }
  }

  const connectWs = () => {
    if (window["WebSocket"]) {
      return new Promise((resolve, reject) => {
        const conn = new WebSocket("ws://" + hostApi + "/ws/" + channel)
        conn.onopen = () => {
          resolve(dispatch(setConn(conn)))
        }
        conn.onclose = (evt) => {
          resolve(dispatch(setMessage(createNotification("Connection Closed."))))
        };
        conn.onmessage = (evt) => {
          const serverLog: ServerLog = JSON.parse(evt.data)
          console.log(serverLog)
          resolve(dispatch(setMessage(serverLog)))
        };
        conn.onerror = function (err) {
          reject(err);
        };
      })
    } else {
      dispatch(setMessage(createNotification("This browser does not support WebSockets.")))
    }
  }

  useEffect(() => {
    (async () => {
      await connectWs()
    })();
  }, [])

  useEffect(() => {
    if (scrollRef && scrollRef.current) {
      const diff = scrollRef.current.scrollHeight - scrollRef.current.clientHeight
      scrollRef.current.scrollTo(0, diff)
    }
  }, [scrollRef, state.msg])

  return (
    <div className={classes.root}>
      <Sidebar users={state.users} />

      <div className={classes.chatBox}>
        <div className={classes.toolbar} />
        <div className={classes.inner}>
          <List className={classes.list} ref={scrollRef}>
            {state.msg.map((msg, index) => (
              <ListItem className={classes.chatItem} key={index} >
                <ListItemAvatar>
                  <Avatar alt="Avatar" />
                </ListItemAvatar>
                <ListItemText
                  primary={
                    <Typography variant="body2" color="textSecondary">
                      {msg.username}
                    </Typography>
                  }
                  secondary={
                    <Typography variant="body1" color="textPrimary">
                      <span dangerouslySetInnerHTML={{ __html: msg.text }} />
                    </Typography>
                  }
                />
              </ListItem>
            ))}
          </List>

          <Grid container alignItems="center" className={classes.sendWrapper}>
            <Grid item xs={10} >
              <TextField
                autoFocus
                label="Message"
                variant="outlined"
                color="secondary"
                fullWidth
                multiline
                rowsMax={2}
                size="small"
                InputLabelProps={{
                  shrink: true,
                }}
                value={state.buffer}
                onChange={(e) => dispatch(setBuffer(e.currentTarget.value))}
                onKeyDown={(e) => enterSendMsg(e)}
              />
            </Grid>
            <Grid item xs={2} >
              <IconButton color="secondary" area-label="send" onClick={sendMsg} >
                <SendIcon />
              </IconButton>
            </Grid>
          </Grid>
        </div>
      </div>
    </div>
  );
}

export default Chat;

const useStyles = makeStyles(theme => ({
  root: {
    display: 'flex',
  },
  chatBox: {
    flexGrow: 1,
    backgroundColor: theme.palette.background.paper,
    margin: theme.spacing(2),
  },
  inner: {
    position: 'relative',
    maxHeight: `calc(100vh - 128px)`,
    height: `calc(100vh - 128px)`,
  },
  list: {
    height: 'calc(100% - 50px)',
    overflowY: 'scroll',
    overflowScrolling: "touch",
  },
  chatContinue: {
    padding: `0 ${theme.spacing(2)}px`,
  },
  chatItem: {
    '&:hover': {
      backgroundColor: theme.palette.action.hover
    },
  },
  sendWrapper: {
    backgroundColor: theme.palette.background.paper,
    position: 'absolute',
    bottom: 0,
    left: 0,
    padding: `0 ${theme.spacing(1)}px`,
    borderTop: '1px solid #ddd',
  },
  toolbar: theme.mixins.toolbar,
}))
