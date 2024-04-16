import React, { useState } from "react";
import "./Login.scss"

interface LoginProps {
    setUser: (user:string) => void
}

const inputs_empty:{[key:string]:string |null} = {
    username: null
}

export const Login = ({setUser}:LoginProps) => {
    const [inputs, setInputs] = useState(inputs_empty)

    const inputsValid = () => {
        let valid:boolean = true
        if(!inputs["username"]){
            valid = false 
            console.error("Invalid username")
        }
        return valid
    }
    const submitInputs = () => {
        if(inputsValid()){
            setUser(inputs["username"]!)
        }
    }
    return (
        <div className="Login">
            <input type="text" name="Username" id="username" placeholder="Insert a username..." 
            onChange={new_value=>{
                setInputs(prevInputs=>{return ({...prevInputs, ["username"]: new_value.target.value})})
            }} />
            <input type="submit" value="Submit" onClick={()=>{submitInputs()}} />
        </div>
    )
}