/**
 * API集成测试工具
 * 用于测试前后端API接口的连通性和正确性
 */

import { http } from '@/utils/http'

interface TestResult {
  name: string
  endpoint: string
  method: string
  status: 'success' | 'error' | 'pending'
  responseTime: number
  statusCode?: number
  error?: string
  data?: any
}

interface TestSuite {
  name: string
  tests: TestCase[]
}

interface TestCase {
  name: string
  endpoint: string
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  headers?: Record<string, string>
  data?: any
  expectedStatus?: number
  validate?: (response: any) => boolean | string
}

class APITester {
  private results: TestResult[] = []
  private isRunning = false

  /**
   * 运行单个测试用例
   */
  async runTest(testCase: TestCase): Promise<TestResult> {
    const startTime = performance.now()
    const result: TestResult = {
      name: testCase.name,
      endpoint: testCase.endpoint,
      method: testCase.method,
      status: 'pending',
      responseTime: 0
    }

    try {
      let response: any

      switch (testCase.method) {
        case 'GET':
          response = await http.get(testCase.endpoint, {
            headers: testCase.headers
          })
          break
        case 'POST':
          response = await http.post(testCase.endpoint, testCase.data, {
            headers: testCase.headers
          })
          break
        case 'PUT':
          response = await http.put(testCase.endpoint, testCase.data, {
            headers: testCase.headers
          })
          break
        case 'DELETE':
          response = await http.delete(testCase.endpoint, {
            headers: testCase.headers
          })
          break
      }

      result.responseTime = performance.now() - startTime
      result.statusCode = response.status
      result.data = response.data

      // 检查状态码
      if (testCase.expectedStatus && response.status !== testCase.expectedStatus) {
        result.status = 'error'
        result.error = `Expected status ${testCase.expectedStatus}, got ${response.status}`
        return result
      }

      // 自定义验证
      if (testCase.validate) {
        const validation = testCase.validate(response.data)
        if (validation !== true) {
          result.status = 'error'
          result.error = typeof validation === 'string' ? validation : 'Validation failed'
          return result
        }
      }

      result.status = 'success'
    } catch (error: any) {
      result.responseTime = performance.now() - startTime
      result.status = 'error'
      result.statusCode = error.response?.status
      result.error = error.message || 'Unknown error'
    }

    return result
  }

  /**
   * 运行测试套件
   */
  async runTestSuite(testSuite: TestSuite): Promise<TestResult[]> {
    this.isRunning = true
    const results: TestResult[] = []

    console.log(`Running test suite: ${testSuite.name}`)

    for (const testCase of testSuite.tests) {
      console.log(`Running test: ${testCase.name}`)
      const result = await this.runTest(testCase)
      results.push(result)
      this.results.push(result)

      // 短暂延迟，避免请求过于频繁
      await new Promise(resolve => setTimeout(resolve, 100))
    }

    this.isRunning = false
    return results
  }

  /**
   * 获取所有测试结果
   */
  getResults(): TestResult[] {
    return [...this.results]
  }

  /**
   * 清空测试结果
   */
  clearResults(): void {
    this.results = []
  }

  /**
   * 获取测试统计
   */
  getStats() {
    const total = this.results.length
    const success = this.results.filter(r => r.status === 'success').length
    const error = this.results.filter(r => r.status === 'error').length
    const avgResponseTime = this.results.reduce((sum, r) => sum + r.responseTime, 0) / total || 0

    return {
      total,
      success,
      error,
      successRate: total > 0 ? (success / total) * 100 : 0,
      avgResponseTime: Math.round(avgResponseTime)
    }
  }

  /**
   * 导出测试报告
   */
  exportReport(): string {
    const stats = this.getStats()
    const timestamp = new Date().toISOString()

    let report = `# API测试报告\n\n`
    report += `**生成时间**: ${timestamp}\n\n`
    report += `## 测试统计\n\n`
    report += `- 总测试数: ${stats.total}\n`
    report += `- 成功: ${stats.success}\n`
    report += `- 失败: ${stats.error}\n`
    report += `- 成功率: ${stats.successRate.toFixed(2)}%\n`
    report += `- 平均响应时间: ${stats.avgResponseTime}ms\n\n`

    report += `## 详细结果\n\n`

    this.results.forEach((result, index) => {
      report += `### ${index + 1}. ${result.name}\n\n`
      report += `- **接口**: ${result.method} ${result.endpoint}\n`
      report += `- **状态**: ${result.status === 'success' ? '✅ 成功' : '❌ 失败'}\n`
      report += `- **响应时间**: ${result.responseTime.toFixed(2)}ms\n`
      
      if (result.statusCode) {
        report += `- **状态码**: ${result.statusCode}\n`
      }
      
      if (result.error) {
        report += `- **错误信息**: ${result.error}\n`
      }
      
      report += `\n`
    })

    return report
  }
}

// 预定义的测试套件
export const testSuites: TestSuite[] = [
  {
    name: 'Django认证服务测试',
    tests: [
      {
        name: '获取验证码',
        endpoint: '/api/auth/captcha',
        method: 'GET',
        expectedStatus: 200,
        validate: (data) => {
          return data && data.captcha_key && data.captcha_image
        }
      },
      {
        name: '用户注册接口',
        endpoint: '/api/auth/register',
        method: 'POST',
        data: {
          username: 'testuser_' + Date.now(),
          email: 'test_' + Date.now() + '@example.com',
          password: 'TestPassword123!',
          captcha_key: 'test_key',
          captcha_value: 'test_value'
        },
        validate: (data) => {
          // 注册可能因为验证码失败，但接口应该正常响应
          return true
        }
      },
      {
        name: '用户登录接口',
        endpoint: '/api/auth/login',
        method: 'POST',
        data: {
          username: 'testuser',
          password: 'testpassword'
        },
        validate: (data) => {
          // 登录可能失败，但接口应该正常响应
          return true
        }
      }
    ]
  },
  {
    name: 'Django商品服务测试',
    tests: [
      {
        name: '获取商品列表',
        endpoint: '/api/products',
        method: 'GET',
        expectedStatus: 200,
        validate: (data) => {
          return Array.isArray(data) || (data && Array.isArray(data.results))
        }
      },
      {
        name: '获取商品分类',
        endpoint: '/api/categories',
        method: 'GET',
        expectedStatus: 200,
        validate: (data) => {
          return Array.isArray(data) || (data && Array.isArray(data.results))
        }
      },
      {
        name: '获取秒杀活动',
        endpoint: '/api/activities',
        method: 'GET',
        expectedStatus: 200,
        validate: (data) => {
          return Array.isArray(data) || (data && Array.isArray(data.results))
        }
      }
    ]
  },
  {
    name: 'Gin秒杀服务测试',
    tests: [
      {
        name: '获取库存信息',
        endpoint: '/seckill/stock/1',
        method: 'GET',
        expectedStatus: 200,
        validate: (data) => {
          return data && typeof data.stock === 'number'
        }
      },
      {
        name: '参与秒杀接口',
        endpoint: '/seckill/participate',
        method: 'POST',
        data: {
          activity_id: 1,
          user_id: 1,
          quantity: 1
        },
        validate: (data) => {
          // 秒杀可能失败，但接口应该正常响应
          return true
        }
      }
    ]
  }
]

// 创建全局实例
export const apiTester = new APITester()

export default apiTester
