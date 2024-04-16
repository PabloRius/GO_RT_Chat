import React from 'react';
import "./App.css"

import Header from './components/Header/Header';
import { Chat } from './components/Chat/Chat';

function App() {

  return (
    <div className="App">
      <Header />
      <Chat />
    </div>
  )
  
}

export default App;
