import React from 'react';
import { useParams } from 'react-router-dom'
import { makeStyles, createStyles, Theme } from '@material-ui/core'
import Drawer from '@material-ui/core/Drawer';
import AppBar from '@material-ui/core/AppBar';
import Toolbar from '@material-ui/core/Toolbar';
import List from '@material-ui/core/List';
import Typography from '@material-ui/core/Typography';
import Divider from '@material-ui/core/Divider';
import ListItem from '@material-ui/core/ListItem';
import ListItemIcon from '@material-ui/core/ListItemIcon';
import ListItemText from '@material-ui/core/ListItemText';
import Avatar from '@material-ui/core/Avatar';
import { User } from '../../type';

const drawerWidth = 240

type SidebarProps = {
  users: User[]
}

const Sidebar: React.FC<SidebarProps> = (props) => {
  const { users } = props
  let { channel } = useParams() as { channel: string }
  const classes = useStyles()

  return (
    <>
      <AppBar position="fixed" className={classes.appBar}>
        <Toolbar>
          <Typography variant="h6" noWrap>
            {`Channel Name: "${channel}"`}
          </Typography>
        </Toolbar>
      </AppBar>
      <Drawer
        className={classes.drawer}
        variant="permanent"
        classes={{
          paper: classes.drawerPaper,
        }}
        anchor="left"
      >
        <div className={classes.toolbar} />
        <Divider />
        <List>
          {users.map((user, index) => (
            <ListItem button key={user.id}>
              <ListItemIcon><Avatar alt={`${user.username} avatar`} /></ListItemIcon>
              <ListItemText primary={user.username} />
            </ListItem>
          ))}
        </List>
      </Drawer>
    </>
  )
}

export default Sidebar

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    appBar: {
      width: `calc(100% - ${drawerWidth}px)`,
      marginLeft: drawerWidth,
    },
    drawer: {
      width: drawerWidth,
      flexShrink: 0,
    },
    drawerPaper: {
      width: drawerWidth,
    },
    toolbar: theme.mixins.toolbar,
  })
)
