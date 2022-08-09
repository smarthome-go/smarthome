<script>
    import Button, { Label } from "@smui/button";
    import Dialog, {
        Actions,
        Content,
        InitialFocus,
        Title,
    } from "@smui/dialog";
    import Textfield from "@smui/textfield";
    import HelperText from "@smui/textfield/helper-text";
    import { createSnackbar } from "../../../global";

    export let open;

    let password = "";
    let passwordConfirm = "";

    let passwordDirty = false;
    let passwordConfirmDirty = false;

    async function submitNewPassword() {
        try {
            const res = await (
                await fetch("/api/user/password/modify", {
                    method: "PUT",
                    body: JSON.stringify({ password }),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            window.location.href = "/logout";
        } catch (err) {
            $createSnackbar(`Failed to update password: ${err}`);
        }
    }
</script>

<Dialog
    bind:open
    aria-labelledby="default-focus-title"
    aria-describedby="default-focus-content"
>
    <Title id="default-focus-title">Update Password</Title>
    <Content id="default-focus-content">
        <Textfield
            bind:dirty={passwordDirty}
            bind:value={password}
            required
            type="password"
            invalid={password === "" && passwordDirty}
            label="Password"
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>Choose a new password</HelperText>
            </svelte:fragment>
        </Textfield>
        <Textfield
            bind:dirty={passwordConfirmDirty}
            bind:value={passwordConfirm}
            required
            type="password"
            invalid={(password !== passwordConfirm || passwordConfirm === "") &&
                passwordConfirmDirty}
            label="Confirm Password"
            style="width: 100%;"
        >
            <svelte:fragment slot="helper">
                <HelperText>Enter your new password again</HelperText>
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
            on:click={submitNewPassword}
            disabled={password === "" || password !== passwordConfirm}
        >
            <Label>Update</Label>
        </Button>
    </Actions>
</Dialog>
