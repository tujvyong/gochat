import React from 'react'
import { makeStyles, createStyles, Theme } from '@material-ui/core'
import { Box } from '@material-ui/core'


const Layout: React.FC = ({ children }) => {
  const classes = useStyles()
  return (
    <Box className={classes.root}>
      {children}
    </Box>
  )
}

export default Layout;

const useStyles = makeStyles((theme: Theme) =>
  createStyles({
    root: {
      position: 'relative',
      backgroundColor: theme.palette.background.default,
      minHeight: '100vh',
      maxWidth: '100vw',
      padding: `${theme.spacing(2)}px 0`,
    },
  })
)
