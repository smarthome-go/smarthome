<script lang="ts">
    import { data, createSnackbar, sleep } from "../../global";
    import Fab, { Icon } from "@smui/fab";
    import ChangeAvatar from "./dialogs/ChangeAvatar.svelte";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import FormField from "@smui/form-field";
    import Switch from "@smui/switch";
    import ColorPicker from "../../components/ColorPicker.svelte";
    import Button from "@smui/button";
    import DeleteAvatar from "./dialogs/DeleteAvatar.svelte";
    import DeleteUser from "./dialogs/DeleteUser.svelte";
    import ChangePassword from "./dialogs/ChangePassword.svelte";

    $: if ($data.userData) receiveInitialData();
    $: if ($data.userData.user.username) reloadAvatarFromSource();

    // Loading indicator
    let loading = false;

    // If forceload is > 0, then load everything againg
    export let forceLoad = true;
    $: if (forceLoad) fetch;

    // User deletion dialog
    let deleteUserOpen = false;

    // Password change dialog
    let changePasswordOpen = false;

    // Avatar-specific values
    let deleteAvatarOpen = false;
    let changeAvatarOpen = false;
    let avatarImageDiv: HTMLDivElement = undefined;

    function changeAvatarSource(src: string) {
        avatarImageDiv.style.backgroundImage = `url(${src})`;
    }
    function reloadAvatarFromSource() {
        avatarImageDiv.style.backgroundImage = `url(/api/user/avatar/personal?urscache=${
            $data.userData.user.username
        }&time=${new Date().getTime()})`;
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

            $data.userData.user.primaryColorDark = primaryColorDark;
            $data.userData.user.primaryColorLight = primaryColorLight;

            // Is required to avoid color bug when switching theme live
            await sleep(100);
            $data.userData.user.darkTheme = darkTheme;
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

    // Deletes the current user
    async function deleteCurrentUser() {
        try {
            const res = await (
                await fetch("/api/user/manage/delete/self", {
                    method: "DELETE",
                })
            ).json();
            if (!res.success) throw Error(res.error);

            // Redirect to the login page
            window.location.href = "/login";
        } catch (err) {
            $createSnackbar(`Failed to delete user: ${err}`);
        }
    }

    let dataReceived = false;

    // Copies the user data from the global store when mounted
    function receiveInitialData() {
        forename = $data.userData.user.forename;
        surname = $data.userData.user.surname;
        schedulerEnabled = $data.userData.user.schedulerEnabled;
        darkTheme = $data.userData.user.darkTheme;
        primaryColorDark = $data.userData.user.primaryColorDark;
        primaryColorLight = $data.userData.user.primaryColorLight;
        setTimeout(() => (dataReceived = true), 100);
    }
</script>

<ChangeAvatar
    bind:open={changeAvatarOpen}
    on:update={(e) => changeAvatarSource(e.detail)}
/>

<DeleteAvatar bind:open={deleteAvatarOpen} on:reset={reloadAvatarFromSource} />
<DeleteUser bind:open={deleteUserOpen} on:delete={deleteCurrentUser} />
<ChangePassword bind:open={changePasswordOpen} />

<div class="preview ">
    <div class="preview__avatar">
        <div class="preview__avatar__image" bind:this={avatarImageDiv} />
        <div class="preview__avatar__edit">
            <Fab
                id="avatar-reset-button"
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
                style="width: 100%;"
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
                style="width: 100%;"
                helperLine$style="width: 100%;"
                label="Surname"
                input$maxlength={20}
                bind:value={surname}
            >
                <CharacterCounter slot="helper">0 / 20</CharacterCounter>
            </Textfield>
        </div>
    </div>
    <div class="inputs__center__wrap">
        <h6>Toggles</h6>
        <div class="inputs__toggles mdc-elevation--z3">
            <div>
                <FormField>
                    <Switch bind:checked={schedulerEnabled} />
                    <span slot="label" class="inputs__toggles__description">
                        Schedules & Automations {schedulerEnabled
                            ? "enabled"
                            : "disabled"}
                    </span>
                </FormField>
            </div>
            <div>
                <FormField>
                    <Switch bind:checked={darkTheme} />
                    <span slot="label" class="inputs__toggles__description">
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
    </div>
    <h6>Danger Zone</h6>
    <div class="inputs__danger mdc-elevation--z3">
        <div class="inputs__danger__item">
            <Button
                class="inputs__danger__item__button"
                on:click={() => (changePasswordOpen = true)}>Change</Button
            >
            <span class="text-hint">Change your Smarthome login password</span>
        </div>
        <div class="inputs__danger__item">
            <Button
                class="inputs__danger__item__button"
                on:click={() => (deleteUserOpen = true)}>Delete</Button
            >
            <span class="text-hint"
                >Erase all your data and delete this account</span
            >
        </div>
    </div>
    <div class="inputs__actions">
        <Button
            on:click={receiveInitialData}
            disabled={!dataReceived ||
                (forename === $data.userData.user.forename &&
                    surname === $data.userData.user.surname &&
                    schedulerEnabled === $data.userData.user.schedulerEnabled &&
                    darkTheme === $data.userData.user.darkTheme &&
                    primaryColorDark === $data.userData.user.primaryColorDark &&
                    primaryColorLight ===
                        $data.userData.user.primaryColorLight)}>Cancel</Button
        >
        <Button
            on:click={updateUserData}
            variant="raised"
            disabled={!dataReceived ||
                (forename === $data.userData.user.forename &&
                    surname === $data.userData.user.surname &&
                    schedulerEnabled === $data.userData.user.schedulerEnabled &&
                    darkTheme === $data.userData.user.darkTheme &&
                    primaryColorDark === $data.userData.user.primaryColorDark &&
                    primaryColorLight ===
                        $data.userData.user.primaryColorLight)}>Apply</Button
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

                @include mobile {
                    height: 5rem;
                }
            }
            &__edit {
                position: absolute;
                right: 0;
                bottom: 0;

                @include mobile {
                    position: relative;
                }

                :global #avatar-reset-button {
                    background-color: var(--clr-error);
                    transform: translateX(calc(100% + 5px)) scale(95%);

                    @include mobile {
                        transform: none;
                    }
                }
                &:hover {
                    :global #avatar-reset-button {
                        transform: translateX(0);
                    }
                }
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

        @include mobile {
            padding: 1.5rem;
        }

        h6 {
            margin-bottom: 0.5rem;
            margin-top: 1rem;
            font-size: 1.1rem;
            color: var(--clr-text-hint);
        }

        &__name {
            display: flex;
            gap: 2rem;

            @include mobile {
                flex-direction: column;
                gap: 0;
            }
        }

        &__toggles {
            background-color: var(--clr-height-1-3);
            border-radius: 0.3rem;
            padding: 1rem;
            display: flex;
            gap: 1rem;
            flex-direction: column;

            :global &__description {
                color: var(--clr-text-hint);
            }
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
                border-radius: 0.3rem;
                padding: 1rem;

                @include not-widescreen {
                    padding: 2rem 1rem;
                    width: 50%;
                }

                @include mobile {
                    width: auto;
                    padding: 1rem;
                }

                div {
                    display: flex;
                    align-items: center;
                    gap: 0.5rem;
                }

                &__indicator {
                    width: 2rem;
                    height: 2rem;
                    border-radius: 50%;

                    @include mobile {
                        width: 1.2rem;
                        height: 1.2rem;
                    }
                }
            }
        }

        &__danger {
            background-color: var(--clr-height-1-3);
            padding: 1.5rem;
            border-radius: 0.3rem;
            display: flex;
            flex-direction: column;
            gap: 1rem;

            @include mobile {
                padding: 0.8rem 1rem;
            }

            &__item {
                display: flex;
                justify-content: space-between;
                align-items: center;

                @include mobile {
                    flex-wrap: wrap;
                    border-left: 0.2rem solid var(--clr-error);
                    padding-left: 0.5rem;
                }

                :global &__button {
                    --mdc-theme-primary: var(--clr-error);
                }
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
