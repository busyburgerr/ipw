<template>
  <div class="about">
    <div class="card">
      <h1>О площадке</h1>
      <p>
        IPW — биржа фриланс-услуг с безопасной сделкой. Заказчик размещает
        проект, фрилансеры отправляют отклики. Принятый оффер превращается в
        контракт с этапами: деньги за этап резервируются в escrow и
        выплачиваются исполнителю только после приёмки работы.
      </p>
      <p>
        Комиссия площадки удерживается при выплате. Спорные ситуации решает
        арбитр — он может вернуть средства заказчику или выплатить их фрилансеру.
      </p>
    </div>
  </div>
</template>

<style scoped>
.about { width: 100%; display: flex; justify-content: center; }
.card {
  background: #fff; color: #000; border-radius: 30px; padding: 50px;
  max-width: 640px; font-family: Inter, sans-serif;
}
.card h1 { font-size: 28px; font-weight: 700; margin-bottom: 16px; }
.card p { font-weight: 500; line-height: 1.6; margin-top: 12px; }
</style>
