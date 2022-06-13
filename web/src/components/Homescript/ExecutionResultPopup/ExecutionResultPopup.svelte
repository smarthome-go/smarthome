<script lang="ts">
    import Dialog, {
        Actions,
        Content,
        Header,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import Button, { Label } from "@smui/button";
    import type { homescriptResponse } from "../../../homescript";
    import { createSnackbar } from "../../../global";
    import { Terminal } from "xterm";
    import { createEventDispatcher, onMount } from "svelte";

    const dispatch = createEventDispatcher()

    // Keeps track of wether the dialog should be open or not
    export let open: boolean = false;
    export let data: homescriptResponse

    onMount(() => {
        let term = new Terminal();
        term.open(document.getElementById("terminal"));
        //  term.write("Hello from \x1B[1;3;31mxterm.js\x1B[0m $ ");
        term.write(data.output)
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
                dispatch('close', null)
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
