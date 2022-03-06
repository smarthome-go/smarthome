const sleep = (ms: number) => new Promise((res) => setTimeout(res, ms));

window.onload = async () => {
  const fileUploadButton = document.getElementById(
    "submit-button"
  ) as HTMLButtonElement;
  fileUploadButton.onclick = async () => {
    const url = "/api/user/avatar/upload";
    const file = document.getElementById("file") as HTMLInputElement;
    let photoList = file.files as FileList;
    if (photoList.length === 0) {
      return 0;
    }
    const photo = photoList[0];
    let formData = new FormData();
    formData.append("file", photo);
    fetch(url, {
      method: "POST",
      body: formData,
    });
  };
  const file = document.getElementById("file") as HTMLInputElement;
  file.onchange = () => {
    let photoList = file.files as FileList;
    if (photoList.length > 0) {
      fileUploadButton.className = "button"
    } else {
      fileUploadButton.className = "button deactivated"
    }
  };
};
