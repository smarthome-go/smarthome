<script lang="ts">
    import Button, { Label } from "@smui/button";

    import Dialog, { Actions, Content, Title } from "@smui/dialog";
    import { createSnackbar } from "../../../global";

    export let open = false;
    $: if (open) fileInput.click();

    let avatarSrc = "";
    let avatarImage: File = undefined;
    let fileInput: HTMLInputElement = undefined;

    const onFileSelected = (e) => {
        avatarImage = e.target.files[0];
        let reader = new FileReader();
        reader.readAsDataURL(avatarImage);
        reader.onload = (e) => {
            avatarSrc = e.target.result as string;
        };
    };

    async function submitImage() {
        const url = "/api/user/avatar/upload";
        let formData = new FormData();
        formData.append("file", avatarImage);

        const res = await (
            await fetch(url, {
                method: "POST",
                body: formData,
            })
        ).json();

        if (!res.success) {
            $createSnackbar(``);
        }
    }
</script>

<Dialog
    bind:open
    aria-labelledby="simple-title"
    aria-describedby="simple-content"
>
    <Title id="simple-title">Dialog Title</Title>
    <Content id="simple-content">
        <img src={avatarSrc} alt="" />
        <input
            style="display:none"
            type="file"
            accept=".jpg, .jpeg, .png"
            on:change={(e) => onFileSelected(e)}
            bind:this={fileInput}
        />
        <Button on:click={submitImage}>submit</Button>
    </Content>
    <Actions>
        <Button>
            <Label>No</Label>
        </Button>
        <Button>
            <Label>Yes</Label>
        </Button>
    </Actions>
</Dialog>
