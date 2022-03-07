window.onload = async () => {
  const avatar = document.getElementById("avatar") as HTMLDivElement;
  avatar.style.backgroundImage = `url('/api/user/avatar?${Date.now()}')`;
}