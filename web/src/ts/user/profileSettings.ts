window.onload = async () => {
  const avatar = document.getElementById("avatar") as HTMLDivElement;
  avatar.style.backgroundImage = `url('/api/user/avatar?${Date.now()}')`;
  const file = document.getElementById("file") as HTMLInputElement;
  file.onchange = async () => {
    showUploadDialog();
  };
};

function showUploadDialog() {
  const div = document.createElement("div");
  div.className = "threeDp avatar-upload-dialog";

  const buttonDiv = document.createElement("div");
  buttonDiv.className = "avatar-button-div";

  const submitButton = document.createElement("button");
  submitButton.className = "button";
  submitButton.innerText = "upload avatar";

  const cancelButton = document.createElement("button");
  cancelButton.className = "button";
  cancelButton.innerText = "cancel";

  const file = document.getElementById("file") as HTMLInputElement;
  let photoList = file.files as FileList;
  if (photoList.length == 0) return;

  const photo = photoList[0];
  const preview = document.createElement("div");
  preview.className = "preview sixDp";

  div.appendChild(preview);
  buttonDiv.appendChild(submitButton);
  buttonDiv.appendChild(cancelButton);
  div.appendChild(buttonDiv);
  document.body.appendChild(div);

  const reader = new FileReader();
  reader.onload = (e) => {
    preview.style.backgroundImage = `url('${e.target?.result as string}')`;
  };
  reader.readAsDataURL(photo);

  cancelButton.onclick = async () => {
    div.style.opacity = "0";
    div.style.marginTop = "-100vh";
    setTimeout(() => {
        div.remove();
    }, 500)
  };

  submitButton.onclick = async () => {
    const url = "/api/user/avatar/upload";
    let formData = new FormData();
    formData.append("file", photo);
    const res = await fetch(url, {
      method: "POST",
      body: formData,
    });
    const body = await res.json();
    if (!body.success) {
      // TODO: proper error handling for frontend, such as a custom error popup
      alert(body.error);
    }
    div.style.opacity = "0";
    div.style.marginTop = "-100vh";
    const avatar = document.getElementById("avatar") as HTMLDivElement;
    avatar.style.backgroundImage = `url('/api/user/avatar?${Date.now()}')`;
    setTimeout(() => {
        div.remove();
    }, 500)
  };
}
