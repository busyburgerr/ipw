<template>
  <div class="container">
    <div class="interactive-textarea">
      <textarea v-model="text" @input="addBlock"></textarea>
      <div v-for="(block, index) in blocks" :key="index" class="block">{{ block }}</div>
    </div>
  </div>
</template>

<script lang="ts">
import {defineComponent} from "vue";

export default defineComponent({
  data() {
    return {
      text: '',
      blocks: [],
    }
  },
  methods: {
    addBlock() {
      const words = this.text.split(' ')
      this.blocks = words
    }
  }
})

</script>

<style scoped>
.container {
  width: 100%;
}
textarea {
  width: 100%;
  height: 200px;
  border: none;
  border-radius: 20px;
  padding: 20px;
  font-size: 16px;
}
.block {
  display: inline-block;
  border: 1px solid black;
  margin: 5px;
}
</style>