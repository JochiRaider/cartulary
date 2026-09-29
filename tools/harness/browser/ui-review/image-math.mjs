export function fraction(count, total) {
  return Number((BigInt(count) * 2000000n + BigInt(total)) / (2n * BigInt(total))) / 1000000;
}
