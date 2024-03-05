import express from "express";
import ViteExpress from "vite-express";
import { handler } from "./build/handler.js";

const app = express();
ViteExpress.config({ mode: "production" })
app.use(handler);
ViteExpress.listen(app, process.env.PORT, () => console.log("Listening on port", process.env.PORT));

process.on("SIGINT", (_) => {
    console.log("Shutting down...");
    setTimeout(shutdown, 1000);
})

process.on("SIGTERM", (_) => {
    console.log("Shutting down...");
    setTimeout(shutdown, 1000);
})

function shutdown() {
    console.log("Bye!");
    process.exit(0);
}
