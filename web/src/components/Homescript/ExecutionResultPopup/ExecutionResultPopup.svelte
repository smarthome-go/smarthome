<script lang="ts">
    import Dialog, { Actions, Content, Header, Title } from "@smui/dialog";
    import Button, { Label } from "@smui/button";
    import type {
        homescriptError,
        homescriptResponse,
    } from "../../../homescript";
    import { Terminal } from "xterm";
    import { createEventDispatcher, onMount } from "svelte";

    const dispatch = createEventDispatcher();

    // Keeps track of wether the dialog should be open or not
    export let open: boolean = false;

    export let data: homescriptResponse;
    export let code: string = ""

    function printError(err: homescriptError, programCode: string): string {
        const lines = programCode.split("\n");
        let line1 = "";
        if (err.location.line > 1) {
            line1 = `\n \x1b[90m${(err.location.line - 1)
                .toString()
                .padStart(3, "0")} | \x1b[0m${lines[err.location.line - 2]}`;
        }
        const line2 = ` \x1b[90m${(err.location.line-1).toString().padStart(3, '0')} | \x1b[0m${lines[err.location.line]}`
        let line3 = "";
        if (err.location.line < lines.length) {
            line3 = `\n \x1b]90m${err.location.line.toString().padStart(3, '0')} | \x1b[0m${lines[err.location.line]}]`
        }

        const marker = `${''.repeat(err.location.column + 6)}\x1b[1;31m^\x1b[0m]`

            return `\x1b[1;36m${err.errorType}${err.location.filename}\x1b[39m at ${err.location.line}:${err.location.column}\x1b[0m` +
                `\n${line1}\n${line2}\n${marker}\n${line3}\n\n\x1b[1;31m${err.message}\x1b[0m\n`
    }

    onMount(() => {
        let term = new Terminal();
        // TODO: bind terminal div to variable instead
        term.open(document.getElementById("terminal"));
        //  term.write("Hello from \x1B[1;3;31mxterm.js\x1B[0m $ ");
        term.write(data.output);
        for (let err of data.error)
            term.write(printError(err, code));
    });
</script>

<Dialog bind:open aria-labelledby="title" aria-describedby="content" fullscreen>
    <Header>
        <Title id="title">Output of ...</Title>
    </Header>
    <Content id="content">
        <div id="terminal" />
    </Content>
    <Actions>
        <Button
            on:click={() => {
                dispatch("close", null);
            }}
        >
            <Label>Close</Label>
        </Button>
    </Actions>
</Dialog>

<style lang="scss">
    #terminal {
        width: 100%;
    }
</style>
