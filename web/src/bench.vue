<template>
  <div class="bench">
    <h3>服务器配置</h3>
    <form class="pure-form pure-form-aligned">
      <fieldset>
        <div class="pure-control-group">
          <label>服务器地址</label>
          <input v-model="serverConfig.url" type="text" />
          <span class="pure-form-message-inline">结尾的 "/" 是必要的</span>
        </div>
        <div class="pure-control-group">
          <label>账号</label>
          <input type="text" v-model="serverConfig.user_id" />
        </div>
        <div class="pure-control-group">
          <label>密码</label>
          <input type="password" v-model="serverConfig.password" />
        </div>
        <div class="pure-control-group">
          <label>测试次数</label>
          <input type="number" v-model="serverConfig.total" />
        </div>
        <div class="pure-control-group">
          <label>预期并发</label>
          <input type="number" v-model="serverConfig.concurrency" />
        </div>
        <div class="pure-control-group">
          <label>题目 ID</label>
          <input type="number" v-model="serverConfig.problem_id" />
        </div>
        <div class="pure-control-group">
          <label>题目解答</label>
          <textarea v-model="serverConfig.answer" />
        </div>
      </fieldset>
    </form>
    <button class="pure-button start" @click="startBench" :disabled="running">
      开始跑分
    </button>
    <h3>跑分结果</h3>
    <div class="result">
      <template v-if="renderList.length">
        <div v-for="item in renderList" :key="item.id" class="result-card">
          <template v-if="item.stage === 'benching' || item.stage === 'end'">
            <div class="result-content">
              <div class="result-count">
                <text class="text-center">结果统计</text>
                <div class="count-item">
                  <text class="count-label">200</text>
                  <text class="count-number">{{ item.count200 }}</text>
                </div>
                <div class="count-item">
                  <text class="count-label">50x</text>
                  <text class="count-number">{{ item.count50x }}</text>
                </div>
                <div class="count-item">
                  <text class="count-label">Unknown</text>
                  <text class="count-number">{{ item.countUnknown }}</text>
                </div>
                <div class="count-item">
                  <text class="count-label">耗时</text>
                  <text class="count-number">{{ item.timecost }}ms</text>
                </div>
              </div>
              <div class="result-performance">
                <div class="performance-progress">
                  <div class="progress-prefix">进度</div>
                  <div class="progress-content">{{ item.progress }}%</div>
                </div>
                <div class="performance-qps">
                  <div class="empty" />
                  <div class="qps-content">{{ item.qps }}</div>
                  <div class="qps-suffix">qps</div>
                </div>
              </div>
            </div>
          </template>
          <template v-else>
            <div class="desc">{{ item.desc }}</div>
          </template>
        </div>
      </template>
      <div v-else class="result-empty">暂无结果</div>
    </div>
  </div>
</template>

<script>
import { reactive, ref, computed } from "vue";

const answer = `#include <stdio.h>
 
int get_sum(int n)
{
    int sum = 0, i;
    for (i = 1; i <= n; i++)
        sum += i;
 
    return sum;
}
 
int main()
{
    int input;
 
    while (scanf("%d", &input) != EOF)
        printf("%d\\n\\n", get_sum(input));
 
    return 0;
}
`;

const benchStage = {
  // 与跑分服务创建任务
  ping: "ping",
  // 检查是否可以登录
  login_check: "login_check",
  // 检查是否在跑分模式下
  check_mode: "check_mode",
  // 跑分中
  benching: "benching",
  end: "end",
};

function wait(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

const socket = new WebSocket("ws://localhost:8000/ws");

const dispatchMap = new Map();

function wsListen(id, cb) {
  dispatchMap.set(id, cb);
  console.log('wsListen', dispatchMap);
}

function wsOff(id) {
  dispatchMap.delete(id);
}

socket.addEventListener("message", function (event) {
  console.log("Message from server ", event.data);
  const data = JSON.parse(event.data);
  const { id } = data;
  if (dispatchMap.has(id)) {
    // console.log("tiger callback");
    dispatchMap.get(id)(data);
  }
});

export default {
  setup() {
    const running = ref(false);
    const resultList = ref([]);
    const serverConfig = reactive({
      url: "https://127.0.0.1/",
      user_id: "admin",
      password: "",
      total: "100",
      concurrency: "20",
      problem_id: "1000",
      answer,
    });
    const handleBenchJob = async (job, _serverConfig) => {
      wsListen(job.id, (wsData) => {
        const { stage, error } = wsData;
        if (error) {
          alert(error);
          wsOff(job.id);
          running.value = false;
          return;
        }
        switch (stage) {
          case "pong": {
            // console.log("pong", job);
            if (job.stage === benchStage.ping) {
              job.stage = benchStage.login_check;
              job.desc = "检查登录有效性";
            }
            break;
          }
          case "login_check": {
            // console.log("login_check", job);
            if (job.stage === benchStage.login_check) {
              job.stage = benchStage.check_mode;
              job.desc = "检查是否在跑分模式下";
            }
            break;
          }
          case "benching": {
            job.stage = benchStage.benching;
            job.count200 = 0;
            job.count50x = 0;
            job.countUnknown = 0;
            job.progress = 0;
            job.timecost = 0;
            job.qps = 0;
            break;
          }
          case "bench_update":
          case "bench_end": {
            if (job.stage !== benchStage.end) {
              const { count200, count50x, countUnknown, timecost } = wsData;
              job.count200 = count200;
              job.count50x = count50x;
              job.countUnknown = countUnknown;
              job.progress =
                (100 * (count200 + count50x + countUnknown)) /
                _serverConfig.total;
              job.timecost = timecost;
              job.qps = (count200 / (timecost / 1000)).toFixed(1);
            }
            if (stage === "bench_end") {
              job.progress = 100;
              job.stage = benchStage.end;
              wsOff(job.id);
              running.value = false;
            }

            break;
          }
        }
      });
      socket.send(
        JSON.stringify({
          action: "start_bench",
          id: job.id,
          payload: _serverConfig,
        })
      );
    };
    const startBench = () => {
      if (running.value) {
        return;
      }
      running.value = true;
      resultList.value.push({
        id: Math.random().toString(),
        stage: benchStage.ping,
        desc: "创建跑分任务",
      });
      handleBenchJob(resultList.value[resultList.value.length - 1], {
        ...serverConfig,
      });
    };
    const renderList = computed(() => resultList.value.reverse());
    return { serverConfig, running, resultList, startBench, renderList };
  },
};
</script>

<style lang="less">
.bench {
  width: 900px;
  margin: 0 auto;

  .start {
    background-color: var(--color-primary);
    color: var(--color-light);
    margin-left: 11em;
  }

  textarea {
    width: 500px;
    height: 500px;
  }

  .text-center {
    text-align: center;
  }

  .result {
    .result-empty,
    .result-card {
      box-shadow: 0 0 4px 0px var(--color-shadow);
      margin-bottom: 20px;
      background-color: var(--color-white);

      transition: box-shadow 0.1s;
    }
    .result-card:hover,
    .result-empty:hover {
      box-shadow: 0 0 20px 0px var(--color-shadow);
    }
    .result-card {
      width: 100%;
      height: 300px;
    }
    .result-empty,
    .desc {
      width: 100%;
      height: 300px;
      line-height: 300px;
      text-align: center;
    }

    .result-content {
      display: flex;
      height: 100%;

      .result-count {
        width: 240px;
        height: 100%;

        display: flex;
        flex-direction: column;
        justify-content: space-evenly;

        .count-item {
          .count-label {
            display: inline-block;
            box-sizing: content-box;
            width: 120px;
            padding-right: 10px;
            text-align: right;
          }
        }
      }

      .result-performance {
        flex: 1;

        .performance-progress,
        .performance-qps {
          height: 150px;
          display: flex;
        }

        .performance-progress {
          padding-left: 100px;
          align-items: flex-end;
        }
        .performance-qps {
          padding-right: 100px;
          align-items: baseline;
        }
        .progress-content,
        .qps-content {
          font-size: 3em;
          line-height: 0.85em;
        }
        .progress-prefix,
        .qps-suffix {
          color: var(--color-secondary);
        }
        .empty {
          flex: 1;
        }
      }
    }
  }
}
</style>
