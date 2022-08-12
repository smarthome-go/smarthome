<script lang="ts">
    import { Icon, Label } from "@smui/button";
    import Button from "@smui/button/src/Button.svelte";
    import Progress from "../../components/Progress.svelte";
    import { createSnackbar } from "../../global";
    import ConfigImport from "./dialogs/ConfigImport.svelte";

    let importConfigOpen = false;

    let loading = false;

    function downloadTextFile(filename: string, content: string) {
        const temp = document.createElement("a");
        temp.href =
            "data:text/plain;charset=utf-8," + encodeURIComponent(content);
        temp.download = filename;
        temp.style.display = "Configuration Export";
        document.body.appendChild(temp);
        temp.click();
        document.body.removeChild(temp);
    }

    async function exportConfig() {
        loading = true;
        try {
            const res = await (await fetch("/api/system/config/export")).json();
            if (res.success != undefined && !res.success)
                throw Error(res.error);
            // Download the fetched configuration
            downloadTextFile(
                `${
                    window.location.hostname
                }_${new Date().toISOString()}_smarthome_export.json`,
                JSON.stringify(res, null, "\t")
            );
        } catch (err) {
            $createSnackbar(`Failed to export system configuration: ${err}`);
        }
        loading = false;
    }
</script>

<ConfigImport bind:open={importConfigOpen} />

<Progress bind:loading />

<div class="container">
    <h6>Export / Import</h6>
    <div class="container__description">
        <span class="text-hint">
            Will export nearly all configured settings of this server.
            <br />
            <strong style="color: var(--clr-error);">Warning!</strong>
            It is strongly recommended to store the export in a safe location, it
            contains hashed passwords and other sensitive information.
        </span>
    </div>

    <br />
    <Button on:click={exportConfig} variant="raised">
        <Label>Export</Label>
        <Icon class="material-icons">file_download</Icon>
    </Button>
    <Button on:click={() => (importConfigOpen = true)} variant="outlined">
        <Label>Import</Label>
        <Icon class="material-icons">file_upload</Icon>
    </Button>
</div>

<style lang="scss">
    .container {
        h6 {
            margin-bottom: 0.5rem;
            margin-top: 1rem;
            font-size: 1.1rem;
            color: var(--clr-text-hint);
        }

        &__description {
            font-size: 0.9rem;
        }
    }
</style>
