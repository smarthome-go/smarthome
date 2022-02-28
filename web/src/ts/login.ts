window.onload = () => {
  const loginbutton = document.getElementById(
    "login-button"
  ) as HTMLButtonElement;
  loginbutton.onclick = async () => {
    await sendLoginRequest();
  };
};

interface LoginRequest {
  username: string;
  password: string;
}

async function sendLoginRequest() {
  const loader = document.getElementById("loader") as HTMLDivElement;
  loader.style.width = "30%";
  const username = document.getElementById("username") as HTMLInputElement;
  const password = document.getElementById("password") as HTMLInputElement;
  const request: LoginRequest = {
    username: username.value,
    password: password.value,
  };
  const loginPostUrl = "/api/login";
  const res = await fetch(loginPostUrl, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify(request),
  });
  switch (res.status) {
    case 204:
      console.log("login success!");
      window.location.href = "/";
      loader.style.width = "100%";
      break;
    case 403:
      console.log("invalid credentials");
      username.style.borderColor = "var(--clr-error)";
      password.style.borderColor = "var(--clr-error)";
      loader.style.backgroundColor = "var(--clr-error)";
      console.log(await res.json());
      setTimeout(() => {
        loader.style.backgroundColor = "var(--clr-primary)";
        username.style.borderColor = "rgb(148, 148, 148)";
        password.style.borderColor = "rgb(148, 148, 148)";
        loader.style.width = "0%";
      }, 1000);
      break;
    case 400:
      console.log("bad request");
      console.log(await res.json());
      break;
    case 500:
      console.log("server error");
      console.log(await res.json());
      break;
  }
}
