import React from 'react';
import { BrowserRouter as Router, Switch, Route } from 'react-router-dom'
import Layout from './component/layout'
import Home from './component/home/home'
import Chat from './component/chats/chat';


const App = () => {
  return (
    <Router>
      <Layout>
        <Switch>
          <Route path="/:channel" component={Chat} />
          <Route path="/" component={Home} />
        </Switch>
      </Layout>
    </Router>
  )
}

export default App
