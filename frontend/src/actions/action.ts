import { ServerLog } from '../type'

export enum ActionType {
  BUFFER = 'BUFFER',
  CONNECTION = 'CONNECTION',
  RECIVED_LOG = 'RECIVED_LOG',
}

export const setBuffer = (newBuffer: string) => ({
  type: ActionType.BUFFER,
  payload: newBuffer,
})
export const setMessage = (newMsg: ServerLog) => ({
  type: ActionType.RECIVED_LOG,
  payload: newMsg,
})
export const setConn = (conn: WebSocket) => ({
  type: ActionType.CONNECTION,
  payload: conn,
})
