<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'

interface Product {
  id: number
  name: string
  price: number
}

interface CartItem {
  product: Product
  quantity: number
}

const products = ref<Product[]>([])
const cart = ref<CartItem[]>([])
const loading = ref(true)
const isCheckingOut = ref(false)
const notification = ref('')

const PRODUCT_API = 'http://localhost:8081/products'
const ORDER_API = 'http://localhost:8082/orders'

const fetchProducts = async () => {
  try {
    const res = await fetch(PRODUCT_API)
    products.value = await res.json()
  } catch (e) {
    console.error('Error fetching products', e)
  }
}

const addToCart = (product: Product) => {
  const existingItem = cart.value.find(item => item.product.id === product.id)
  if (existingItem) {
    existingItem.quantity++
  } else {
    cart.value.push({ product, quantity: 1 })
  }
  showNotification(`Added ${product.name} to cart!`)
}

const removeFromCart = (index: number) => {
  cart.value.splice(index, 1)
}

const updateQuantity = (index: number, change: number) => {
  const item = cart.value[index]
  if (item.quantity + change > 0) {
    item.quantity += change
  } else {
    removeFromCart(index)
  }
}

const cartTotal = computed(() => {
  return cart.value.reduce((total, item) => total + (item.product.price * item.quantity), 0)
})

const checkout = async () => {
  if (cart.value.length === 0) return
  
  isCheckingOut.value = true
  try {
    const promises = cart.value.map(item => {
      const orderData = {
        product_id: item.product.id,
        quantity: item.quantity,
        total: item.product.price * item.quantity
      }
      return fetch(ORDER_API, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(orderData)
      })
    })
    
    await Promise.all(promises)
    showNotification('Checkout successful! Your order has been placed.')
    cart.value = [] // clear cart
  } catch (e) {
    console.error('Error during checkout', e)
    showNotification('Error during checkout. Please try again.')
  } finally {
    isCheckingOut.value = false
  }
}

const showNotification = (msg: string) => {
  notification.value = msg
  setTimeout(() => { notification.value = '' }, 3000)
}

onMounted(async () => {
  loading.value = true
  await fetchProducts()
  loading.value = false
})
</script>

<template>
  <div class="p-8">
    <!-- Notification Toast -->
    <div 
      v-if="notification"
      class="fixed top-20 left-1/2 -translate-x-1/2 z-50 px-6 py-3 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 backdrop-blur-xl shadow-[0_0_20px_rgba(16,185,129,0.15)] animate-in fade-in slide-in-from-top-4 duration-300"
    >
      {{ notification }}
    </div>

    <main class="max-w-7xl mx-auto grid grid-cols-1 lg:grid-cols-3 gap-8">
      
      <!-- Products Section -->
      <section class="lg:col-span-2 space-y-6">
        <h2 class="text-3xl font-semibold flex items-center gap-3">
          <span class="w-10 h-10 rounded-xl bg-indigo-500/20 flex items-center justify-center text-indigo-400">
            <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 11V7a4 4 0 00-8 0v4M5 9h14l1 12H4L5 9z"></path></svg>
          </span>
          Featured Products
        </h2>

        <div v-if="loading" class="grid grid-cols-1 md:grid-cols-2 gap-6 mt-8">
          <div v-for="i in 4" :key="i" class="h-64 rounded-3xl bg-slate-800/50 animate-pulse border border-slate-700/50"></div>
        </div>

        <div v-else-if="products.length === 0" class="p-12 mt-8 rounded-3xl bg-slate-800/20 border border-dashed border-slate-700 text-center text-slate-400">
          No products available. Check back later!
        </div>

        <div v-else class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6 mt-8">
          <div 
            v-for="product in products" 
            :key="product.id"
            class="group relative p-8 rounded-3xl bg-slate-800/40 border border-slate-700/50 hover:bg-slate-800/60 transition-all duration-300 hover:-translate-y-2 hover:shadow-[0_20px_40px_rgb(0,0,0,0.2)] hover:border-indigo-500/30 overflow-hidden flex flex-col justify-between h-64"
          >
            <!-- Decorative gradient orb -->
            <div class="absolute -top-12 -right-12 w-40 h-40 bg-indigo-500/10 rounded-full blur-3xl group-hover:bg-indigo-500/20 transition-colors"></div>
            
            <div class="relative z-10 flex flex-col h-full justify-between">
              <div>
                <h3 class="text-2xl font-bold text-slate-100 group-hover:text-indigo-300 transition-colors">{{ product.name }}</h3>
                <p class="text-4xl font-light text-slate-300 mt-3">${{ product.price.toFixed(2) }}</p>
              </div>
              
              <button 
                @click="addToCart(product)"
                class="w-full py-4 px-6 rounded-2xl bg-gradient-to-r from-indigo-500 to-cyan-500 text-white font-semibold text-lg hover:from-indigo-400 hover:to-cyan-400 focus:ring-4 focus:ring-indigo-500/20 transition-all duration-200 shadow-xl shadow-indigo-500/20 active:scale-[0.98] flex items-center justify-center gap-2"
              >
                <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
                <span>Add to Cart</span>
              </button>
            </div>
          </div>
        </div>
      </section>

      <!-- Cart Section -->
      <aside class="space-y-6">
        <h2 class="text-3xl font-semibold flex items-center gap-3">
          <span class="w-10 h-10 rounded-xl bg-cyan-500/20 flex items-center justify-center text-cyan-400">
            <svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
          </span>
          Your Cart
        </h2>

        <div class="p-6 rounded-3xl bg-slate-800/40 border border-slate-700/50 flex flex-col h-[calc(100vh-12rem)] max-h-[800px]">
          
          <div v-if="cart.length === 0" class="flex-1 flex flex-col items-center justify-center text-slate-400 space-y-4">
            <svg class="w-16 h-16 text-slate-600" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"></path></svg>
            <p class="text-lg">Your cart is empty.</p>
          </div>

          <div v-else class="flex-1 overflow-y-auto space-y-4 pr-2 custom-scrollbar">
            <div 
              v-for="(item, index) in cart" 
              :key="item.product.id"
              class="p-4 rounded-2xl bg-slate-800 border border-slate-700 flex justify-between items-center"
            >
              <div>
                <h4 class="font-semibold text-slate-200">{{ item.product.name }}</h4>
                <p class="text-indigo-300 font-medium">${{ item.product.price.toFixed(2) }}</p>
              </div>
              <div class="flex items-center gap-3 bg-slate-900 rounded-xl p-1">
                <button @click="updateQuantity(index, -1)" class="w-8 h-8 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 flex items-center justify-center transition-colors">
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 12H4"></path></svg>
                </button>
                <span class="w-6 text-center text-slate-200 font-medium">{{ item.quantity }}</span>
                <button @click="updateQuantity(index, 1)" class="w-8 h-8 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 flex items-center justify-center transition-colors">
                  <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"></path></svg>
                </button>
              </div>
            </div>
          </div>

          <div v-if="cart.length > 0" class="pt-6 mt-6 border-t border-slate-700 space-y-6">
            <div class="flex justify-between items-end">
              <span class="text-slate-400 font-medium">Total Price</span>
              <span class="text-4xl font-bold text-slate-100">${{ cartTotal.toFixed(2) }}</span>
            </div>
            <button 
              @click="checkout"
              :disabled="isCheckingOut"
              class="w-full py-4 rounded-2xl bg-gradient-to-r from-emerald-500 to-teal-500 text-white font-bold text-xl hover:from-emerald-400 hover:to-teal-400 focus:ring-4 focus:ring-emerald-500/20 shadow-xl shadow-emerald-500/20 active:scale-[0.98] transition-all disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
            >
              <svg v-if="!isCheckingOut" class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
              <svg v-else class="w-6 h-6 animate-spin" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"></path></svg>
              <span>{{ isCheckingOut ? 'Processing...' : 'Checkout Now' }}</span>
            </button>
          </div>

        </div>
      </aside>

    </main>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(30, 41, 59, 0.5);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(71, 85, 105, 0.8);
  border-radius: 10px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(100, 116, 139, 1);
}
</style>
