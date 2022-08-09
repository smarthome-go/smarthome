<script>
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import { createEventDispatcher } from "svelte";
    import { createSnackbar } from "../../../global";

    export let open = false;
    export let token = "";

    // Event dispatcher
    const dispatch = createEventDispatcher();

    async function deleteToken() {
        try {
            const res = await (
                await fetch("/api/user/token/delete", {
                    method: "DELETE",
                    body: JSON.stringify({ token }),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            dispatch("delete", null);
        } catch (err) {
            $createSnackbar(`Failed to delete authentication token: ${err}`);
        }
    }
</script>

<Dialog
    bind:open
    aria-labelledby="default-focus-title"
    aria-describedby="default-focus-content"
>
    <Title id="default-focus-title">Confirm Deletion</Title>
    <Content id="default-focus-content">
        <span class="text-hint">
            You are about to delete an authentication token. This means you will
            not be able to login on clients using this token.
        </span>
    </Content>
    <Actions>
        <Button on:click={deleteToken}>
            <Label>Delete</Label>
        </Button>
        <Button defaultAction use={[InitialFocus]}>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>
