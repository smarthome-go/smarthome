<script lang="ts">
    import Button, { Icon, Label } from "@smui/button";
    import Dialog, { Actions, Content, Header, Title } from "@smui/dialog";

    export let open = false;

    let fileInput: HTMLInputElement = undefined;
    let importedJSONString = "";

    // Callback to be executed as soon as a file has been picked
    function onFileSelected(e: Event): string {
        const importFile = (e.target as HTMLInputElement).files[0];
        if (importFile === undefined) {
            open = false;
            return;
        }
        const reader = new FileReader();
        reader.readAsText(importFile);
        reader.onload = (e) => {
            importedJSONString = reader.result as string;
        };
    }
</script>

<input
    style="display:none"
    type="file"
    accept=".json"
    on:input={(e) => onFileSelected(e)}
    bind:this={fileInput}
/>

<Dialog bind:open fullscreen aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">Configuration Import</Title>
    </Header>
    <Content id="content">
        <textarea id="preview">{importedJSONString}</textarea>
        <Button on:click={() => fileInput.click()}>
            <Label>Upload</Label>
        </Button>
    </Content>
    <Actions>
        <Button>
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style>
    #preview {
        font-family: "Jetbrains Mono", monospace;
        outline: none;
        height: 30rem;
        width: 100%;
        resize:none;
    }
</style>
