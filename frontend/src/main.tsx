import {createRoot} from 'react-dom/client'
import { HashRouter } from 'react-router-dom'
import './style.css'
import App from './App'

const container = document.getElementById('root')

const root = createRoot(container!)

root.render(
    <HashRouter>
        <App/>
    </HashRouter>
)
