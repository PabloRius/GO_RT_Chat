import React, { useEffect, useState } from "react";
import useWebSocket, { ReadyState } from "react-use-websocket";
import Axios from "axios"
import "./Chat.scss"

import { Login } from "../Login/Login";
import { mock_chat_data } from "../../mock/chat_data";

interface Message{
    sender?: string,
    recipient?: string,
    content: string
}

const inputs_empty:{[key:string]:string |null} = {
    message: null
}

export const Chat = () => {
    const [user, setUser] = useState<string | null>(null)
    const [inputs, setInputs] = useState(inputs_empty)
    const [history, setHistory] = useState<any[]>([])
    const [receiver, setReceiver] = useState<null | string>(null)
    const [chats, setChats] = useState<any[]>(mock_chat_data)

    const WS_URL = `ws://localhost:12345/ws?username=${user}`
    const { sendJsonMessage, lastJsonMessage } = useWebSocket(
        WS_URL,
        {
            share: false,
            shouldReconnect: () => true,
            onMessage: (event) => {fetchHistory()},
        }
    )
    useEffect(()=>{
        if(user) {
            Axios.post(`http://localhost:12345/chats?username=${user}`, {
            }).then((response:any)=>{setChats(response.data)})
        }
    },[user])
    useEffect(()=>{
        if (receiver) {
            Axios.post(`http://localhost:12345/history?username=${user}&receiver=${receiver}`, {
            }).then((response:any)=>{setHistory(response.data)})
        }
    },[receiver])
    const fetchHistory = () => {
        if (receiver) {
            Axios.post(`http://localhost:12345/history?username=${user}&receiver=${receiver}`, {
            }).then((response:any)=>{setHistory(response.data)})
        }
    }

    const sendMessage = (message:string) => {
        sendJsonMessage({
            recipient: receiver,
            sender: user,
            content: message
        })
    }
    
    const inputsValid = () => {
        let valid:boolean = true
        if(!inputs["message"]){
            valid = false 
            console.error("Invalid message")
        }
        return valid
    }

    const submitInputs = () => {
        if(inputsValid()){
            sendMessage(inputs["message"]!)
        }
    }

    if (user){
        return (
            <div className="Chat">
                <div className="Active">
                    {chats && chats.map((chat)=>(
                        <div className={`ActiveChat ${chat.username === receiver ? "active": ""}`} onClick={()=>{setReceiver(chat.username)}}>
                            <p>{chat.username!}</p>
                        </div>
                    ))}
                </div>
                <div className="Body">
                    <div className="History">
                        {history && history.map((message)=>(
                            <p className={`${message.sender === user ? "sender": ""}`}>{message.content || ""}</p>
                            ))}
                    </div>
                    <form>
                        <input type="textarea" name="Message" id="message" placeholder="Write a message..." disabled={receiver ? false: true}
                        onChange={new_value=>{
                            setInputs(prevInputs=>{return ({...prevInputs, ["message"]: new_value.target.value})})
                        }} />
                        <input type="submit" disabled={receiver ? false: true} value="Submit" onClick={(event)=>{event.preventDefault(); submitInputs(); fetchHistory()}} />
                    </form>
                </div>
            </div>
        )
    }
    return <Login setUser={setUser} />
}
