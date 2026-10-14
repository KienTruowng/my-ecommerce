
<template>
  <div class="min-h-screen grid lg:grid-cols-2 bg-slate-50">
    <!-- Left side -->
    <section
      class="hidden lg:flex flex-col justify-between p-10 text-white"
      style="background: linear-gradient(135deg, #07172d, #0b2347, #175cd3);"
    >
      <div class="text-2xl font-bold tracking-wider text-amber-300">
        KIEN STORE
      </div>

      <div class="space-y-5">
        <p class="text-sm font-semibold tracking-widest text-amber-300">
          PREMIUM SHOPPING EXPERIENCE
        </p>

        <h1 class="text-5xl font-bold leading-tight">
          Discover
          <br />
          Extraordinary.
        </h1>

        <p class="max-w-md text-base leading-7 text-blue-100">
          Chào mừng đến với KIEN STORE.
          Khám phá những sản phẩm chất lượng
          và trải nghiệm mua sắm cao cấp.
        </p>
      </div>

      <div class="text-xs text-blue-100">
        KIEN STORE - Premium Shopping
      </div>
    </section>

    <!-- Right side -->
    <section class="flex items-center justify-center p-6 sm:p-10">
      <div class="w-full max-w-md rounded-3xl border border-slate-200 bg-white p-8 shadow-xl">
        <slot />
      </div>
    </section>
  </div>
</template>
