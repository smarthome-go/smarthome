<script lang="ts">
    import { data, createSnackbar } from "../../global";
    import Fab, { Icon } from "@smui/fab";
    import ChangeAvatar from "./dialogs/ChangeAvatar.svelte";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import FormField from "@smui/form-field";
    import Switch from "@smui/switch";
    import ColorPicker from "../../components/ColorPicker.svelte";
    import Button from "@smui/button";
    import DeleteAvatar from "./dialogs/DeleteAvatar.svelte";

    $: if ($data.userData) receiveInitialData();

    // Loading indicator
    let loading = false;

    // Avatar-specific values
    let deleteAvatarOpen = false;

    let changeAvatarOpen = false;
    let avatarImageDiv: HTMLDivElement = undefined;

    function changeAvatarSource(src: string) {
        avatarImageDiv.style.backgroundImage = `url(${src})`;
    }
    function reloadAvatarFromSource() {
        avatarImageDiv.style.backgroundImage = `url(/api/user/avatar/personal?time=${new Date().getTime()})`;
    }

    // User data copy
    let forename = "";
    let surname = "";

    let schedulerEnabled = false;
    let darkTheme = false;

    let primaryColorDark = "";
    let primaryColorLight = "";

    async function updateUserData() {
        loading = true;
        try {
            // Update regular data
            const res = await (
                await fetch("/api/user/data/update", {
                    method: "PUT",
                    headers: { "Content-Type": "appliation/json" },
                    body: JSON.stringify({
                        forename,
                        surname,
                        primaryColorDark,
                        primaryColorLight,
                    }),
                })
            ).json();
            if (!res.success) throw Error(res.error);

            // Update the scheduler state afterwards
            await setScheduler();

            // Update the theme afterwards
            await setTheme();

            // If everything until now succeeded, update the values in the global store
            $data.userData.user.forename = forename;
            $data.userData.user.surname = surname;

            $data.userData.user.schedulerEnabled = schedulerEnabled;
            $data.userData.user.darkTheme = darkTheme;

            $data.userData.user.primaryColorDark = primaryColorDark;
            $data.userData.user.primaryColorLight = primaryColorLight;
        } catch (err) {
            $createSnackbar(`Failed to update user data: ${err}`);
        }
        loading = false;
    }

    // Toggles the users scheduler
    async function setScheduler() {
        if (schedulerEnabled == $data.userData.user.schedulerEnabled) return;
        try {
            const res = await (
                await fetch("/api/scheduler/state/personal", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        enabled: schedulerEnabled,
                    }),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            // Update the value in the global store
            $data.userData.user.schedulerEnabled = true;
        } catch (err) {
            throw Error(err);
        }
    }

    // Toggles the users theme preference
    async function setTheme() {
        if (darkTheme == $data.userData.user.darkTheme) return;
        try {
            const res = await (
                await fetch("/api/user/settings/theme/personal", {
                    method: "PUT",
                    headers: { "Content-Type": "application/json" },
                    body: JSON.stringify({
                        darkTheme,
                    }),
                })
            ).json();
            if (!res.success) throw Error(res.error);
            // Update value in theme
            $data.userData.user.darkTheme = darkTheme;
        } catch (err) {
            throw Error(err);
        }
    }

    // Copies the user data from the global store when mounted
    function receiveInitialData() {
        forename = $data.userData.user.forename;
        surname = $data.userData.user.surname;
        schedulerEnabled = $data.userData.user.schedulerEnabled;
        darkTheme = $data.userData.user.darkTheme;
        primaryColorDark = $data.userData.user.primaryColorDark;
        primaryColorLight = $data.userData.user.primaryColorLight;
    }
</script>

<ChangeAvatar
    bind:open={changeAvatarOpen}
    on:update={(e) => changeAvatarSource(e.detail)}
/>

<DeleteAvatar bind:open={deleteAvatarOpen} on:reset={reloadAvatarFromSource} />

<div class="preview ">
    <div class="preview__avatar">
        <div class="preview__avatar__image" bind:this={avatarImageDiv} />
        <div class="preview__avatar__edit">
            <Fab
                color="primary"
                on:click={() => (deleteAvatarOpen = true)}
                mini
            >
                <Icon class="material-icons">delete</Icon>
            </Fab>
            <Fab
                color="primary"
                on:click={() => (changeAvatarOpen = true)}
                mini
            >
                <Icon class="material-icons">edit</Icon>
            </Fab>
        </div>
    </div>
    <div class="preview__name">
        <span class="preview__name__fullname">
            {$data.userData.user.forename}
            {$data.userData.user.surname}
        </span>
        <span class="preview__name__username">
            {$data.userData.user.username}
        </span>
    </div>
</div>
<div class="inputs">
    <div class="inputs__name">
        <div class="inputs__name__forename">
            <Textfield
                helperLine$style="width: 100%;"
                label="Forename"
                input$maxlength={20}
                bind:value={forename}
            >
                <CharacterCounter slot="helper">0 / 20</CharacterCounter>
            </Textfield>
        </div>
        <div class="inputs__name__surname">
            <Textfield
                helperLine$style="width: 100%;"
                label="Surname"
                input$maxlength={20}
                bind:value={surname}
            >
                <CharacterCounter slot="helper">0 / 20</CharacterCounter>
            </Textfield>
        </div>
    </div>
    <h6>Toggles</h6>
    <div class="inputs__toggles mdc-elevation--z3">
        <div>
            <FormField>
                <Switch bind:checked={schedulerEnabled} />
                <span slot="label">
                    Schedules & Automations {schedulerEnabled
                        ? "enabled"
                        : "disabled"}
                </span>
            </FormField>
        </div>
        <div>
            <FormField>
                <Switch bind:checked={darkTheme} />
                <span slot="label">
                    Darkmode {darkTheme ? "enabled" : "disabled"}
                </span>
            </FormField>
        </div>
    </div>
    <h6>Primary Colors</h6>
    <div class="inputs__primary-colors">
        <div class="color mdc-elevation--z3">
            <div
                class="color__indicator"
                style:background-color={primaryColorDark}
            />
            <!-- Primary Color Dark -->
            <div>
                <ColorPicker bind:value={primaryColorDark} />
                <span>Dark</span>
            </div>
        </div>
        <div class="color mdc-elevation--z3">
            <div
                class="color__indicator"
                style:background-color={primaryColorLight}
            />
            <!-- Primary Color Light -->
            <div>
                <ColorPicker bind:value={primaryColorLight} />
                <span>Light</span>
            </div>
        </div>
    </div>
    <h6 style="color: var(--clr-error)">Danger Zone</h6>
    <div class="inputs__danger mdc-elevation--z3">
        <div class="inputs__danger__delete-user">
            <Button variant="outlined">Delete</Button>
            <div>
                <span class="--clr-text-hint"
                    >Erase all your data and delete this account</span
                >
            </div>
        </div>
    </div>
    <div class="inputs__actions">
        <Button on:click={receiveInitialData}>Cancel</Button>
        <Button
            on:click={updateUserData}
            disabled={forename === $data.userData.user.forename &&
                surname === $data.userData.user.surname &&
                schedulerEnabled === $data.userData.user.schedulerEnabled &&
                darkTheme === $data.userData.user.darkTheme &&
                primaryColorDark === $data.userData.user.primaryColorDark &&
                primaryColorLight === $data.userData.user.primaryColorLight}
            >Apply Changes</Button
        >
    </div>
</div>

<style lang="scss">
    @use "../../mixins" as *;
    .preview {
        padding: 1rem 1.5rem;
        display: flex;
        gap: 2rem;

        &__avatar {
            position: relative;

            &__image {
                background-position: center;
                background-size: cover;
                background-repeat: no-repeat;
                border-radius: 50%;
                aspect-ratio: 1;
                height: 8rem;
                background-image: url("/api/user/avatar/personal");
            }
            &__edit {
                position: absolute;
                right: 0;
                bottom: 0;
            }
        }

        &__name {
            display: flex;
            flex-direction: column;
            justify-content: center;

            &__fullname {
                font-weight: bold;
                font-size: 1.5rem;
            }
            &__username {
                color: var(--clr-text-hint);
                font-size: 1.2rem;
                font-family: monospace;
            }
        }
    }
    .inputs {
        padding: 1rem 2rem;

        h6 {
            margin-bottom: 0.5rem;
            margin-top: 1rem;
        }

        &__name {
            display: flex;
            gap: 2rem;
        }

        &__toggles {
            background-color: var(--clr-height-1-3);
            border-radius: 0.3rem;
            padding: 1rem;
            display: flex;
            gap: 1rem;
            flex-direction: column;
        }

        &__primary-colors {
            display: flex;
            gap: 1rem;

            @include mobile {
                flex-direction: column;
            }

            .color {
                display: flex;
                align-items: center;
                gap: 0.2rem;
                background-color: var(--clr-height-1-3);
                padding: 1rem;
                border-radius: 0.3rem;

                div {
                    display: flex;
                    align-items: center;
                    gap: 0.5rem;
                }

                &__indicator {
                    width: 2rem;
                    height: 2rem;
                    border-radius: 50%;

                    @include not-widescreen {
                        width: 1.2rem;
                        height: 1.2rem;
                    }
                }
            }
        }

        &__danger {
            background-color: var(--clr-height-1-3);
            padding: 1rem;
            border-radius: 0.3rem;
            border: var(--clr-error) solid 0.1rem;
            display: flex;
            flex-direction: column;
            gap: 1rem;

            div {
                display: flex;
                justify-content: space-between;
            }
        }

        &__actions {
            display: flex;
            justify-content: flex-end;
            gap: 0.5rem;
            margin-top: 1.5rem;
        }
    }
</style>
