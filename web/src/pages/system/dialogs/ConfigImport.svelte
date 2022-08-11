<script lang="ts">
    import Button, { Icon, Label } from "@smui/button";
    import Checkbox from "@smui/checkbox";
    import Dialog, { Actions, Content, Header, Title } from "@smui/dialog";
    import FormField from "@smui/form-field";
    import Progress from "../../../components/Progress.svelte";
    import { createSnackbar, sleep } from "../../../global";

    export let open = false;

    let confirm = false;

    let importStarted = false;
    let importRunning = false;

    let statusMessage = "";
    let statusError = "";

    let remainingSecs = 0;

    async function redirect() {
        for (remainingSecs = 5; remainingSecs > 0; remainingSecs--)
            await sleep(1000);
        window.location.href = "/logout";
    }

    async function importConfig() {
        importRunning = true;
        importStarted = true;
        try {
            const res = await (
                await fetch("/api/system/config/import", {
                    method: "POST",
                    headers: { "Content-Type": "application/json" },
                    body: importedJSONString,
                })
            ).json();
            if (!res.success) {
                statusMessage = res.message;
                statusError = res.error;
            } else {
                redirect();
            }
        } catch (err) {
            $createSnackbar(`Could not import configuaration: ${err}`);
        }
        importRunning = false;
    }

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

<Dialog bind:open aria-labelledby="title" aria-describedby="content">
    <Header>
        <Title id="title">Configuration Import</Title>
    </Header>
    <Content id="content">
        {#if importedJSONString === ""}
            <div id="upload">
                <span
                    >Upload a smarhome config export file to get started.
                </span>
                <div class="list">
                    <span> How it works: </span>
                    <ul>
                        <li>
                            Before the import, the server is reset to factory
                            defaults
                        </li>
                        <li>
                            When the file is imported, it's contents are
                            evaluated
                        </li>
                        <li>You are notified about the outcome</li>
                    </ul>
                </div>
                <Button variant="raised" on:click={() => fileInput.click()}>
                    <Label>Upload</Label>
                    <Icon class="material-icons">file_upload</Icon>
                </Button>
            </div>
        {:else if importStarted}
            <Progress bind:loading={importRunning} />
            {#if statusMessage !== "" || statusError !== ""}
                <span id="import-output">
                    {statusMessage}
                    <br />
                    <span style="color: var(--clr-error)">
                        {statusError}
                    </span>
                </span>
            {:else}
                <span style="color: var(--clr-success)">
                    <strong>Success: </strong>
                    the new configuration has been successfully imported.
                </span>
                <span class="text-hint">
                    You will be redirected to login in {remainingSecs} second(s)
                </span>
            {/if}
        {:else}
            <div id="confirm">
                <div class="list warn">
                    <span>Before you continue</span>
                    <ul>
                        <li>
                            Importing will <strong>erase all data</strong> on this
                            server
                        </li>
                        <li>
                            Do <strong>not shutdown</strong> the server during the
                            process
                        </li>
                        <li>Proceeding comes at your own risk</li>
                        <li>Importing may take some time</li>
                    </ul>
                </div>
                <FormField>
                    <Checkbox id="import-confirm" bind:checked={confirm} />
                    <span slot="label"
                        >I known the risk and want to continue.</span
                    >
                </FormField>
                <Button
                    id="import-button"
                    variant="raised"
                    disabled={importedJSONString === "" || !confirm}
                    on:click={importConfig}
                >
                    <Label>Import</Label>
                </Button>
            </div>
        {/if}
    </Content>
    <Actions>
        <Button
            on:click={() => {
                importedJSONString = "";
                confirm = false;
            }}
        >
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    @use "../../../mixins" as *;

    #upload {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        gap: 1rem;

        width: 30rem;

        @include mobile {
            width: auto;
        }
    }

    #confirm {
        width: 30rem;

        @include mobile {
            width: auto;
        }
    }

    .list {
        margin-bottom: 0.5rem;

        &.warn {
            span {
                color: var(--clr-error);
            }
        }

        span {
            font-weight: bold;
            color: var(--clr-primary);
        }

        ul {
            padding: 0 1rem;
            margin: 0;
            margin-top: 0.125rem;
        }
    }
    :global #import-confirm {
        --mdc-theme-secondary: var(--clr-error);
        --mdc-ripple-color: var(--clr-error);
    }
    :global #import-button {
        --mdc-theme-primary: var(--clr-error);
        margin-top: 0.8rem;
        display: block;
    }
    #import-output {
        font-family: "Jetbrains Mono", monospace;
        font-size: 0.9rem;
    }
</style>
