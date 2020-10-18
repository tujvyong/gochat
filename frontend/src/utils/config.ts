// fake Load Blancing
export const hostApi = Date.now() % 2 === 0 ? "localhost:8080" : "localhost:8081"
