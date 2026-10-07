<template>
  <div class="auth-research-illus" aria-hidden="true">
    <svg
      class="auth-research-illus-svg"
      viewBox="0 0 480 260"
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
    >
      <!-- soft floor shadow -->
      <ellipse cx="240" cy="248" rx="168" ry="10" fill="rgba(0,0,0,0.12)" />

      <!-- desk -->
      <rect x="32" y="198" width="416" height="14" rx="5" fill="rgba(255,255,255,0.2)" />
      <rect x="48" y="206" width="384" height="8" rx="3" fill="rgba(255,255,255,0.1)" />

      <!-- keyboard -->
      <rect x="188" y="192" width="104" height="14" rx="5" fill="rgba(255,255,255,0.28)" />
      <rect x="196" y="196" width="12" height="3" rx="1" fill="rgba(255,255,255,0.45)" />
      <rect x="212" y="196" width="12" height="3" rx="1" fill="rgba(255,255,255,0.45)" />
      <rect x="228" y="196" width="12" height="3" rx="1" fill="rgba(255,255,255,0.45)" />
      <rect x="244" y="196" width="12" height="3" rx="1" fill="rgba(255,255,255,0.45)" />
      <rect x="260" y="196" width="12" height="3" rx="1" fill="rgba(255,255,255,0.45)" />

      <!-- paper stack (left) -->
      <g transform="translate(72 172)">
        <rect width="44" height="32" rx="4" fill="rgba(255,255,255,0.35)" />
        <rect x="4" y="-4" width="44" height="32" rx="4" fill="rgba(255,255,255,0.5)" />
        <rect x="8" y="-8" width="44" height="32" rx="4" fill="rgba(255,255,255,0.72)" />
        <rect x="16" y="8" width="26" height="2.5" rx="1.25" fill="#c4b5fd" />
        <rect x="16" y="14" width="22" height="2.5" rx="1.25" fill="#ddd6fe" />
        <rect x="16" y="20" width="28" height="2.5" rx="1.25" fill="#ddd6fe" />
      </g>

      <!-- teacup on saucer (right) -->
      <ellipse cx="368" cy="202" rx="26" ry="5" fill="rgba(255,255,255,0.28)" />
      <path
        d="M348 198c0-14 12-22 20-22s20 8 20 22v2c0 2-2 4-4 4h-32c-2 0-4-2-4-4v-2z"
        fill="rgba(255,255,255,0.75)"
      />
      <ellipse cx="368" cy="176" rx="18" ry="5" fill="rgba(255,255,255,0.55)" />
      <ellipse cx="368" cy="176" rx="14" ry="3.5" fill="#c4b5fd" opacity="0.85" />
      <path
        d="M388 182c6 0 10 4 10 10s-4 10-10 10"
        stroke="rgba(255,255,255,0.7)"
        stroke-width="2.5"
        stroke-linecap="round"
        fill="none"
      />
      <path
        d="M362 168c0-4 3-7 6-7"
        stroke="rgba(255,255,255,0.45)"
        stroke-width="2"
        stroke-linecap="round"
      />
      <path
        d="M368 164c0-5 3-8 6-8"
        stroke="rgba(255,255,255,0.35)"
        stroke-width="2"
        stroke-linecap="round"
      />
      <path
        d="M374 166c0-4 2-6 5-6"
        stroke="rgba(255,255,255,0.4)"
        stroke-width="2"
        stroke-linecap="round"
      />

      <!-- monitor + stand -->
      <rect x="118" y="28" width="244" height="132" rx="12" fill="rgba(255,255,255,0.96)" />
      <rect x="130" y="40" width="220" height="104" rx="8" fill="#f5f3ff" />
      <rect x="142" y="52" width="88" height="7" rx="3.5" fill="#7c3aed" />
      <rect x="142" y="66" width="196" height="4" rx="2" fill="#c4b5fd" />
      <rect x="142" y="76" width="180" height="4" rx="2" fill="#ddd6fe" />
      <rect x="142" y="86" width="188" height="4" rx="2" fill="#ddd6fe" />
      <rect x="142" y="96" width="160" height="4" rx="2" fill="#ddd6fe" />
      <rect x="142" y="106" width="120" height="4" rx="2" fill="#ede9fe" />
      <rect x="142" y="122" width="56" height="10" rx="5" fill="#6366f1" opacity="0.35" />
      <path d="M228 160h24v18h-24z" fill="rgba(255,255,255,0.5)" />
      <rect x="204" y="176" width="72" height="7" rx="3.5" fill="rgba(255,255,255,0.35)" />
    </svg>
  </div>
</template>

<style scoped>
.auth-research-illus {
  display: flex;
  flex: 1;
  align-items: flex-end;
  justify-content: center;
  min-height: 120px;
  margin-top: 4px;
  padding-top: 8px;
  pointer-events: none;
}

.auth-research-illus-svg {
  width: 100%;
  max-width: 400px;
  height: auto;
  opacity: 0.96;
}

@media (max-width: 960px) {
  .auth-research-illus {
    display: none;
  }
}
</style>
