<script>
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import HelperText from "@smui/textfield/helper-text";
    import { createEventDispatcher } from "svelte";
    import { createSnackbar } from "../../../global";

    export let open;
    let label = "";

    // Event dispatcher
    const dispatch = createEventDispatcher();

    async function generateToken() {
        try {
            const res = await (
                await fetch("/api/user/token/generate", {
                    method: "POST",
                    body: JSON.stringify({ label }),
                })
            ).json();
            if (!res.response.success) throw Error(res.response.error);

            // Dispatch the create event to allow the parent to display the change
            dispatch("create", { label, token: res.token });

            // Reset the label on successful submit
            label = ""
        } catch (err) {
            $createSnackbar(
                `Failed to generate new authentication token: ${err}`
            );
        }
    }
</script>

<Dialog
    bind:open
    aria-labelledby="default-focus-title"
    aria-describedby="default-focus-content"
>
    <Title id="default-focus-title">Generate Authentication Token</Title>
    <Content id="default-focus-content">
        <Textfield
            bind:value={label}
            label="Token Name"
            input$maxlength={50}
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>Client name / Description</HelperText>
                <CharacterCounter>0 / 50</CharacterCounter>
            </svelte:fragment>
        </Textfield>
    </Content>
    <Actions>
        <Button>
            <Label>Cancel</Label>
        </Button>
        <Button
            defaultAction
            use={[InitialFocus]}
            on:click={generateToken}
            disabled={label === ""}
        >
            <Label>Generate</Label>
        </Button>
    </Actions>
</Dialog>
