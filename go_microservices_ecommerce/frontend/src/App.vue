<script setup lang="ts">
</script>

<template>
  <div class="min-h-screen bg-slate-900 text-slate-100 font-sans selection:bg-indigo-500/30">
    <!-- Navbar -->
    <nav class="sticky top-0 z-40 w-full backdrop-blur-xl bg-slate-900/80 border-b border-slate-800">
      <div class="max-w-6xl mx-auto px-8 h-16 flex items-center justify-between">
        <router-link to="/" class="text-2xl font-extrabold bg-clip-text text-transparent bg-gradient-to-r from-indigo-400 to-cyan-400 tracking-tight">
          Go Ecommerce Store
        </router-link>
        <div class="flex gap-6 items-center">
          <router-link to="/" class="text-sm font-medium text-slate-300 hover:text-white transition-colors" active-class="text-indigo-400">Shop</router-link>
          <router-link to="/admin" class="px-4 py-2 rounded-lg bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 hover:bg-indigo-500/20 transition-colors text-sm font-semibold" active-class="bg-indigo-500 text-white border-indigo-500">Admin Panel</router-link>
        </div>
      </div>
    </nav>

    <!-- Main View -->
    <router-view v-slot="{ Component }">
      <transition name="fade" mode="out-in">
        <component :is="Component" />
      </transition>
    </router-view>
  </div>
</template>

<style>
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}
.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(5px);
}
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: rgba(30, 41, 59, 0.5);
  border-radius: 8px;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: rgba(71, 85, 105, 0.8);
  border-radius: 8px;
}
.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: rgba(100, 116, 139, 1);
}
</style>
