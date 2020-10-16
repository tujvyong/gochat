export interface User {
  id: string,
  username: string,
}

export type ServerLog = {
  command: string,
  msg: Message,
}

export type Message = {
  username: string,
  text: string,
}
