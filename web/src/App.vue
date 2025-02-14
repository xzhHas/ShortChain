<template>
  <div id="app">
    <div class="container">
      <h1>短链接生成器</h1>
      <p class="description">
        输入长链接，快速生成一个可分享的短链接。
      </p>
      <div class="form-container">
        <label for="long-url" class="label">请输入长链接：</label>
        <input id="long-url" v-model="longUrl" type="text" placeholder="请输入长链接" />
        <button @click="generateShortUrl">生成短链接</button>
      </div>

      <div v-if="shortUrl" class="result-container">
        <p>生成的短链接：</p>
        <div class="short-url-box">
          <input id="short-url" readonly :value="shortUrl" />
        </div>
        <p>
          <a :href="shortUrl" target="_blank" class="short-url-link">访问短链接</a>
        </p>
      </div>
    </div>
  </div>
</template>

<script>
import axios from "axios";

export default {
  data() {
    return {
      longUrl: "",
      shortUrl: "",
    };
  },
  methods: {
    async generateShortUrl() {
      if (!this.longUrl) {
        alert("请输入长链接！");
        return;
      }
      try {
        const response = await axios.post(
          "http://s.golangcode.cn/api/shorten",
          { long_url: this.longUrl },
          {
            headers: {
              "Content-Type": "application/json",
            },
          }
        );
        this.shortUrl = response.data.short_url;
      } catch (error) {
        console.error("生成短链接失败：", error);
        alert("短链接生成失败，请稍后再试！");
      }
    },
    copyToClipboard() {
      navigator.clipboard.writeText(this.shortUrl);
      alert("短链接已复制到剪贴板！");
    },
  },
};
</script>

<style>
/* 全局样式 */
body {
  margin: 0;
  padding: 0;
  font-family: "Arial", sans-serif;
  background: linear-gradient(135deg, #6a11cb, #2575fc);
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 100vh;
  color: #333;
}

/* 容器样式 */
.container {
  background: #fff;
  border-radius: 12px;
  padding: 40px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  text-align: center;
  max-width: 500px;
  width: 80%;
  /* 容器宽度占屏幕的80% */
  animation: fadeIn 0.5s ease-in-out;
}

/* 标题和描述 */
h1 {
  font-size: 2rem;
  margin-bottom: 10px;
  color: #444;
}

.description {
  font-size: 1rem;
  color: #020202;
  margin-bottom: 20px;
}

/* 表单样式 */
.label {
  font-size: 1rem;
  color: #555;
  margin-bottom: 10px;
  display: block;
}

input {
  width: 100%;
  /* 让输入框和显示框宽度一致 */
  padding: 15px;
  font-size: 1rem;
  border: 1px solid #ccc;
  border-radius: 8px;
  margin-bottom: 20px;
  box-sizing: border-box;
}

input:focus {
  outline: none;
  border-color: #6a11cb;
  box-shadow: 0 0 5px rgba(106, 17, 203, 0.5);
}

button {
  width: 100%;
  padding: 15px;
  font-size: 1rem;
  font-weight: bold;
  color: #fff;
  background: linear-gradient(135deg, #4aec80, #1e7e34);
  border: none;
  border-radius: 8px;
  cursor: pointer;
  transition: transform 0.2s ease;
}

button:hover {
  transform: scale(1.05);
}

button:active {
  transform: scale(1);
}

/* 短链接结果样式 */
.result-container {
  margin-top: 30px;
  text-align: center;
}

.short-url-box {
  display: flex;
  justify-content: center;
  align-items: center;
  margin-top: 20px;
}

.short-url-box input {
  width: 100%;
  /* 让输入框和显示框宽度一致 */
  padding: 20px;
  border: 1px solid #ccc;
  border-radius: 8px;
  box-sizing: border-box;
  text-align: center;
}

.short-url-link {
  color: #007bff;
  text-decoration: none;
}

.short-url-link:hover {
  text-decoration: underline;
}

/* 动画效果 */
@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(-20px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
