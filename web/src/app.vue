<template>
  <div>
    <header>
      <div :class="{ selected: menu === 'bench' }" @click="selectMenu('bench')">
        跑分
      </div>
      <div class="divider"></div>
      <div
        :class="{ selected: menu === 'config' }"
        @click="selectMenu('config')"
      >
        配置
      </div>
    </header>
    <config v-if="menu === 'config'" />
    <bench v-if="menu === 'bench'" />
    <footer>
      <text
        >© 2020
        <a href="https://github.com/muzea" target="__blank">muzea</a>.</text
      >
    </footer>
  </div>
</template>

<script>
import { ref } from "vue";
import config from "./config.vue";
import bench from "./bench.vue";

export default {
  components: {
    config,
    bench,
  },
  setup() {
    const menu = ref("bench");
    const selectMenu = (next) => {
      menu.value = next;
    };
    return { menu, selectMenu };
  },
};
</script>

<style lang="less">
body {
  min-height: 100vh;
  background-color: var(--color-light);
}
#app {
  background-color: var(--color-light);
  padding: 10px 0;

  header {
    display: flex;
    align-items: center;
    justify-content: center;

    .divider {
      width: 40px;
    }
    & > * {
      display: inline-block;
      font-size: 2em;
      position: relative;
      z-index: 2;
      cursor: pointer;
      color: var(--color-secondary);
      transition: color 0.3s;

      &::after {
        content: "";
        display: inline-block;
        position: absolute;
        height: 0;
        width: 120%;
        bottom: 0;
        left: -10%;
        z-index: -1;
        background-color: var(--color-warning);

        transition: height 0.3s;
      }
      &.selected {
        color: var(--color-dark);
      }
      &.selected::after {
        height: 10px;
      }
      &:hover {
        color: var(--color-dark);
      }
      &:hover::after {
        height: 20px;
      }
    }
  }

  footer {
    background-color: var(--color-light);
    display: block;
    padding: 10px;
    text-align: center;

    a {
      color: #000;
      text-decoration: dotted;
    }
  }
}
</style>
