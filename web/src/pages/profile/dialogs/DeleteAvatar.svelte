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

    export let open;

    // Event dispatcher
    const dispatch = createEventDispatcher();

    async function resetAvatar() {
        try {
            const res = await (
                await fetch("/api/user/avatar/delete", {
                    method: "DELETE",
                })
            ).json();
            if (!res.success) throw Error(res.error);
            dispatch('reset', null)
        } catch (err) {
            $createSnackbar(`Failed to reset avatar image: ${err}`);
        }
    }
</script>

<Dialog
    bind:open
    aria-labelledby="default-focus-title"
    aria-describedby="default-focus-content"
>
    <Title id="default-focus-title">Confirm Reset</Title>
    <Content id="default-focus-content">
            <span class="text-hint">
                <strong>Note: </strong>
                It might take up to
                <span style="color: var(--clr-primary);">6 hours</span> for the default
                image to appear on every device. If you are impatient, try clearing
                your browser's cache or force-reloading this page.
            </span>
    </Content>
    <Actions>
        <Button on:click={resetAvatar}>
            <Label>Confirm</Label>
        </Button>
        <Button defaultAction use={[InitialFocus]}>
            <Label>Cancel</Label>
        </Button>
    </Actions>
</Dialog>
