<template>
  <div class="min-h-screen grid lg:grid-cols-2">
    <section
      class="hidden lg:flex flex-col justify-between p-10 text-white bg-gradient-to-br from-emerald-600 via-teal-600 to-cyan-700"
    >
      <div class="text-2xl font-semibold tracking-tight">CampusTrade</div>
      <div class="space-y-4">
        <h1 class="text-4xl font-semibold leading-tight">
          Trade smarter.
          <br />
          Meet faster.
        </h1>
        <p class="text-sm text-white/80 max-w-sm">
          Exchange items inside your university community with verified profiles and safe
          hand-offs.
        </p>
      </div>
      <div class="text-xs text-white/70">Powered by your campus network.</div>
    </section>
    <section class="flex items-center justify-center p-6 sm:p-10">
      <div class="w-full max-w-md bg-card/80 border rounded-3xl shadow-xl p-8 backdrop-blur">
        <slot />
      </div>
    </section>
  </div>
</template>
