## Go Real Time Chat
A chat web project
## About the project
This project consists on a web page running in the **React** framework, the components are built using the **typescript** language, the backend is built using the **Go** language, and the messages are distributed using the **weboscket** technology.

### Contents
The main page features a form in which you shall introduce a username 

Once introduced a username, the messages page will appear, on it you can select any prevously opened conversation and send / receive messages with that user 

## Available Scripts

In the frontend directory, you can:
### Run the frontend
The dependencies must be installed first, to do so:
```bash
npm install
```
The dependencies form the included package.json will be installed, once it's done you can run the applcation with:
```bash
npm start
```

Runs the app in the development mode.\
Open [http://localhost:3000](http://localhost:3000) to view it in your preferred browser.

The page will reload if you make edits.\
You will also see any lint errors in the console.

### Build the frontend 
```bash
npm run build
```

Builds the app for production to the `build` folder.\
It correctly bundles React in production mode and optimizes the build for the best performance.

In the server directory you can:

## Run the server
```bash
go run .
```
Once the server is up, the messages service will be available for any client 

## Learn More

### Tech stack used
<div style="display: flex; flex-direction:row; column-gap: 10px">
    <img src="https://upload.wikimedia.org/wikipedia/commons/thumb/a/a7/React-icon.svg/2300px-React-icon.svg.png" height="50px" width="50px" />
    <img src="https://static-00.iconduck.com/assets.00/typescript-icon-icon-1024x1024-vh3pfez8.png" height="50px" width="50px" />
    <img src="https://www.svgrepo.com/show/373632/go.svg" height="50px" width="50px" />
</div>