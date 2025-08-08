export default function Logo({ theme = "dark", className }: { theme?: string, className?: string }) {
  const color = theme === "dark" ? "#ffffff" : "#000000";
  return (
    <svg width="800px" height="800px" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg" className={className}>
      <path d="M 21.894 11.553 C 19.736 7.236 15.904 5 12 5 c -3.903 0 -7.736 2.236 -9.894 6.553 a 1 1 0 0 0 0 0.894 C 4.264 16.764 8.096 19 12 19 c 3.903 0 7.736 -2.236 9.894 -6.553 a 1 1 0 0 0 0 -0.894 z M 12 17 c -2.969 0 -6.002 -1.62 -7.87 -5 C 5.998 8.62 9.03 7 12 7 c 2.969 0 6.002 1.62 7.87 5 c -1.868 3.38 -4.901 5 -7.87 5 z M 11 18 l 0 3 l 2 0 l 0 -3 M 7 17 l 0 3 l 2 0 l 0 -3 M 15 17 l 0 3 l 2 0 l 0 -3 M 12 20 A 1 1 0 0 1 12 23 A 1 1 0 0 1 12 20 M 8 19 A 1 1 0 0 1 8 22 A 1 1 0 0 1 8 19 M 16 19 A 1 1 0 0 1 16 22 A 1 1 0 0 1 16 19 M 12 20 A 1 1 0 0 1 12 23 A 1 1 0 0 1 12 20 M 8 19 A 1 1 0 0 1 8 22 A 1 1 0 0 1 8 19 M 16 19 A 1 1 0 0 1 16 22 A 1 1 0 0 1 16 19" fill={color} />
    </svg>
  )
}
