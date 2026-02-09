import { createApp } from "vue";
import App from "./App.vue";
import { router } from "./router";

/* import the necessary styles for Vue Flow to work */
import "@vue-flow/core/dist/style.css";
import "@vue-flow/core/dist/theme-default.css";
import "./assets/vue-flow-dark.css";

const app = createApp(App);
app.use(router);
app.mount("#app");
