/**
 * Polyfills for older browsers
 * 为旧浏览器提供兼容性支持
 */

/**
 * Promise polyfill for IE
 */
if (!window.Promise) {
  // 简单的Promise polyfill
  window.Promise = class SimplePromise {
    private state: 'pending' | 'fulfilled' | 'rejected' = 'pending'
    private value: any = undefined
    private handlers: Array<{
      onFulfilled?: (value: any) => any
      onRejected?: (reason: any) => any
      resolve: (value: any) => void
      reject: (reason: any) => void
    }> = []

    constructor(executor: (resolve: (value: any) => void, reject: (reason: any) => void) => void) {
      try {
        executor(this.resolve.bind(this), this.reject.bind(this))
      } catch (error) {
        this.reject(error)
      }
    }

    private resolve(value: any) {
      if (this.state === 'pending') {
        this.state = 'fulfilled'
        this.value = value
        this.handlers.forEach(handler => this.handle(handler))
        this.handlers = []
      }
    }

    private reject(reason: any) {
      if (this.state === 'pending') {
        this.state = 'rejected'
        this.value = reason
        this.handlers.forEach(handler => this.handle(handler))
        this.handlers = []
      }
    }

    private handle(handler: any) {
      if (this.state === 'pending') {
        this.handlers.push(handler)
      } else {
        if (this.state === 'fulfilled' && handler.onFulfilled) {
          handler.onFulfilled(this.value)
        }
        if (this.state === 'rejected' && handler.onRejected) {
          handler.onRejected(this.value)
        }
      }
    }

    then(onFulfilled?: (value: any) => any, onRejected?: (reason: any) => any) {
      return new (window.Promise as any)((resolve: any, reject: any) => {
        this.handle({
          onFulfilled: onFulfilled ? (value: any) => {
            try {
              resolve(onFulfilled(value))
            } catch (error) {
              reject(error)
            }
          } : resolve,
          onRejected: onRejected ? (reason: any) => {
            try {
              resolve(onRejected(reason))
            } catch (error) {
              reject(error)
            }
          } : reject,
          resolve,
          reject
        })
      })
    }

    catch(onRejected: (reason: any) => any) {
      return this.then(undefined, onRejected)
    }

    static resolve(value: any) {
      return new (window.Promise as any)((resolve: any) => resolve(value))
    }

    static reject(reason: any) {
      return new (window.Promise as any)((_resolve: any, reject: any) => reject(reason))
    }

    static all(promises: any[]) {
      return new (window.Promise as any)((resolve: any, reject: any) => {
        if (promises.length === 0) {
          resolve([])
          return
        }

        let remaining = promises.length
        const results: any[] = []

        promises.forEach((promise, index) => {
          (window.Promise as any).resolve(promise).then((value: any) => {
            results[index] = value
            remaining--
            if (remaining === 0) {
              resolve(results)
            }
          }).catch(reject)
        })
      })
    }
  } as any
}

/**
 * Fetch polyfill for IE
 */
if (!window.fetch) {
  window.fetch = function(input: URL | RequestInfo, init: RequestInit = {}) {
    const url = typeof input === 'string' ? input : input.toString()
    const options = init as any
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      const method = options.method || 'GET'
      const headers = options.headers || {}
      
      xhr.open(method, url)
      
      // 设置请求头
      Object.keys(headers).forEach(key => {
        xhr.setRequestHeader(key, headers[key])
      })
      
      xhr.onload = function() {
        const response = {
          ok: xhr.status >= 200 && xhr.status < 300,
          status: xhr.status,
          statusText: xhr.statusText,
          headers: new Map(),
          json: () => Promise.resolve(JSON.parse(xhr.responseText)),
          text: () => Promise.resolve(xhr.responseText),
          blob: () => Promise.resolve(new Blob([xhr.response])),
          arrayBuffer: () => Promise.resolve(xhr.response)
        }
        
        if (response.ok) {
          resolve(response as any)
        } else {
          reject(new Error(`HTTP ${xhr.status}: ${xhr.statusText}`))
        }
      }
      
      xhr.onerror = function() {
        reject(new Error('Network error'))
      }
      
      xhr.ontimeout = function() {
        reject(new Error('Request timeout'))
      }
      
      // 发送请求
      if (options.body) {
        xhr.send(options.body)
      } else {
        xhr.send()
      }
    })
  }
}

/**
 * Object.assign polyfill
 */
if (!Object.assign) {
  Object.assign = function(target: any, ...sources: any[]) {
    if (target == null) {
      throw new TypeError('Cannot convert undefined or null to object')
    }
    
    const to = Object(target)
    
    sources.forEach(source => {
      if (source != null) {
        Object.keys(source).forEach(key => {
          to[key] = source[key]
        })
      }
    })
    
    return to
  }
}

/**
 * Array.from polyfill
 */
if (!Array.from) {
  Array.from = function(arrayLike: any, mapFn?: (value: any, index: number) => any, thisArg?: any) {
    const items = Object(arrayLike)
    const len = parseInt(items.length) || 0
    const result = new Array(len)
    
    for (let i = 0; i < len; i++) {
      const value = items[i]
      result[i] = mapFn ? mapFn.call(thisArg, value, i) : value
    }
    
    return result
  }
}

/**
 * Array.includes polyfill
 */
if (!Array.prototype.includes) {
  Array.prototype.includes = function(searchElement: any, fromIndex: number = 0) {
    const len = this.length
    let index = Math.max(fromIndex >= 0 ? fromIndex : len + fromIndex, 0)
    
    while (index < len) {
      if (this[index] === searchElement) {
        return true
      }
      index++
    }
    
    return false
  }
}

/**
 * String.includes polyfill
 */
if (!String.prototype.includes) {
  String.prototype.includes = function(searchString: string, position: number = 0) {
    return this.indexOf(searchString, position) !== -1
  }
}

/**
 * String.startsWith polyfill
 */
if (!String.prototype.startsWith) {
  String.prototype.startsWith = function(searchString: string, position: number = 0) {
    return this.substring(position, position + searchString.length) === searchString
  }
}

/**
 * String.endsWith polyfill
 */
if (!String.prototype.endsWith) {
  String.prototype.endsWith = function(searchString: string, length?: number) {
    const actualLength = length !== undefined ? length : this.length
    const start = actualLength - searchString.length
    return start >= 0 && this.substring(start, start + searchString.length) === searchString
  }
}

/**
 * Element.closest polyfill
 */
if (!Element.prototype.closest) {
  Element.prototype.closest = function(selector: string) {
    let element: Element | null = this
    
    while (element && element.nodeType === 1) {
      if (element.matches && element.matches(selector)) {
        return element
      }
      element = element.parentElement
    }
    
    return null
  }
}

/**
 * Element.matches polyfill
 */
if (!Element.prototype.matches) {
  Element.prototype.matches = 
    (Element.prototype as any).matchesSelector ||
    (Element.prototype as any).mozMatchesSelector ||
    (Element.prototype as any).msMatchesSelector ||
    (Element.prototype as any).oMatchesSelector ||
    (Element.prototype as any).webkitMatchesSelector ||
    function(this: Element, selector: string) {
      const matches = (this.ownerDocument || document).querySelectorAll(selector)
      let i = matches.length
      while (--i >= 0 && matches.item(i) !== this) {}
      return i > -1
    }
}

/**
 * CustomEvent polyfill for IE
 */
if (!window.CustomEvent) {
  window.CustomEvent = function(event: string, params: any = {}) {
    const evt = document.createEvent('CustomEvent')
    // Using deprecated method for IE compatibility
    ;(evt as any).initCustomEvent(event, params.bubbles || false, params.cancelable || false, params.detail || null)
    return evt
  } as any
}

/**
 * 初始化所有polyfills
 */
export function initPolyfills() {
  // 所有polyfills在模块加载时自动执行
  console.log('Polyfills initialized for browser compatibility')
}
