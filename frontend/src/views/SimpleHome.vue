<template>
  <div class="simple-home">
    <div class="hero-section">
      <h1 class="title">🚀 Flash Sku</h1>
      <p class="subtitle">高性能秒杀抢购平台</p>
      <div class="features">
        <div class="feature-card">
          <h3>⚡ 高性能</h3>
          <p>支持大规模并发秒杀</p>
        </div>
        <div class="feature-card">
          <h3>📱 PWA支持</h3>
          <p>可安装的Web应用</p>
        </div>
        <div class="feature-card">
          <h3>🎯 实时更新</h3>
          <p>库存实时同步</p>
        </div>
      </div>
      <div class="actions">
        <button class="btn-primary" @click="goToLogin">立即登录</button>
        <button class="btn-secondary" @click="goToRegister">注册账号</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'

const router = useRouter()

const goToLogin = () => {
  router.push('/auth/login')
}

const goToRegister = () => {
  router.push('/auth/register')
}
</script>

<style scoped>
.simple-home {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  text-align: center;
  padding: 20px;
  margin: 0;
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Hiragino Sans GB', 'Microsoft YaHei', sans-serif;
}

.hero-section {
  max-width: 900px;
  width: 100%;
  padding: 2rem;
}

.title {
  font-size: 3.2rem;
  font-weight: 700;
  margin-bottom: 1.5rem;
  text-shadow: 2px 2px 8px rgba(0,0,0,0.3);
  letter-spacing: -0.02em;
  line-height: 1.2;
}

.subtitle {
  font-size: 1.3rem;
  margin-bottom: 4rem;
  opacity: 0.95;
  font-weight: 300;
  letter-spacing: 0.5px;
  line-height: 1.4;
}

.features {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 1.5rem;
  margin-bottom: 3rem;
  max-width: 700px;
  margin-left: auto;
  margin-right: auto;
}

.feature-card {
  background: rgba(255, 255, 255, 0.15);
  padding: 1.8rem 1.5rem;
  border-radius: 16px;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.25);
  transition: all 0.3s ease;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}

.feature-card:hover {
  transform: translateY(-5px);
  background: rgba(255, 255, 255, 0.2);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.15);
}

.feature-card h3 {
  font-size: 1.4rem;
  margin-bottom: 0.8rem;
  font-weight: 600;
  line-height: 1.3;
}

.feature-card p {
  opacity: 0.9;
  font-size: 0.95rem;
  line-height: 1.5;
  margin: 0;
}

.actions {
  display: flex;
  gap: 1.2rem;
  justify-content: center;
  flex-wrap: wrap;
  margin-top: 2rem;
}

.btn-primary, .btn-secondary {
  padding: 14px 32px;
  border: none;
  border-radius: 12px;
  font-size: 1.05rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  min-width: 140px;
  font-family: inherit;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.btn-primary {
  background: linear-gradient(135deg, #ff6b6b 0%, #ff5252 100%);
  color: white;
  border: 2px solid transparent;
}

.btn-primary:hover {
  background: linear-gradient(135deg, #ff5252 0%, #ff4444 100%);
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(255, 107, 107, 0.4);
}

.btn-secondary {
  background: rgba(255, 255, 255, 0.1);
  color: white;
  border: 2px solid rgba(255, 255, 255, 0.6);
  backdrop-filter: blur(10px);
}

.btn-secondary:hover {
  background: rgba(255, 255, 255, 0.95);
  color: #667eea;
  transform: translateY(-3px);
  box-shadow: 0 8px 24px rgba(255, 255, 255, 0.3);
}

@media (max-width: 768px) {
  .simple-home {
    padding: 15px;
  }

  .hero-section {
    padding: 1.5rem;
  }

  .title {
    font-size: 2.8rem;
    margin-bottom: 1rem;
  }

  .subtitle {
    font-size: 1.1rem;
    margin-bottom: 3rem;
  }

  .features {
    grid-template-columns: 1fr;
    gap: 1.2rem;
    margin-bottom: 2.5rem;
  }

  .feature-card {
    padding: 1.5rem 1.2rem;
  }

  .actions {
    flex-direction: column;
    align-items: center;
    gap: 1rem;
  }

  .btn-primary, .btn-secondary {
    width: 100%;
    max-width: 280px;
    padding: 16px 24px;
  }
}
</style>
