<template>
  <div class="container">
    <div class="container-head">
      <div class="user-block">
        <div class="user-data">
          <div class="user-photo"></div>
          <div class="user-info">
            <div class="nameTag-block">
              <div class="name">Donkey Lover</div>
              <div class="tag">@donkeylover</div>
            </div>
            <div class="email">worker@gmail.com</div>
          </div>
        </div>
        <div class="user-message">Привет, я Donkey Lover!</div>
      </div>
      <div class="user-block">
        <div class="user-data">
          <div class="user-photo"></div>
          <div class="user-info">
            <div class="nameTag-block">
              <div class="name">Donkey Lover</div>
              <div class="tag">@donkeylover</div>
            </div>
            <div class="email">worker@gmail.com</div>
          </div>
        </div>
        <div class="user-message">Привет, я Donkey Lover!</div>
      </div>
    </div>
    <div class="container-content">
      <div class="msg-head">
        <div class="user-data">
          <div class="user-photo"></div>
          <div class="user-info">
            <div class="nameTag-block">
              <div class="name">Donkey Lover</div>
              <div class="tag">@donkeylover</div>
            </div>
            <div class="email">worker@gmail.com</div>
          </div>
        </div>
      </div>
      <div class="msg-content">
        <div class="message">
          <div class="other-message_block">
            <div class="other-message">Привет, я Donkey Lover!</div>
          </div>
        </div>
        <div class="message">
          <div class="my-message_block">
            <div class="my-message">Да поебать мне</div>
          </div>
        </div>
        <div class="message">
          <div class="my-message_block">
<!--            <div class="my-message">Да поебать мне</div>-->
            <div class="my-message" v-for="(message, index) in messages" :key="index">
              {{ message }}
            </div>
          </div>
        </div>
      </div>
      <div class="msg-footer">
        <input v-model="inputMessage" />
        <button type="submit" @click="sendMessage">Отправить</button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import axios from "axios";

const userData = ref([])
export default {
  data() {
    return {
      messages: [],
      inputMessage: '',
      inputFrom: '',
      inputTo: '',
      ws: null,
    };
  },
  setup() {
    onMounted(async () => {
      try {
        const res = await axios.get(`http://localhost:5000/data/v1/user`, {withCredentials: true})
        userData.value = res.data
        console.log()
      } catch (e) {
        console.log(e)
      }
    })
  },
  mounted() {
    this.setupWebSocket();
  },
  methods: {
    setupWebSocket() {
      this.ws = new WebSocket('ws://localhost:5000/message');
      this.ws.onopen = () => {
        console.log('websocket connect...');
      };
      this.ws.onmessage = (event) => {
        const messageObject = JSON.parse(event.data);
        this.messages.push(messageObject.message);
        console.log("messageObject.from:", messageObject.from);
        console.log("messageObject.to:", messageObject.to);
        console.log("messageObject.message:", messageObject.message);
      };
    },
    sendMessage() {
      const messageObject = {
        from: userData.value['id'],
        to: Number(this.inputTo),
        message: this.inputMessage,
      };
      this.ws.send(JSON.stringify(messageObject));
      this.inputMessage = '';
      console.log("message", messageObject.message);
    },
  },
};
</script>

<style scoped>
@import "./assets/css/style.css";

* {
  color: #000;
}
</style>