<script lang="ts">
    import { data } from "../../global";
    import Fab, { Icon } from "@smui/fab";
    import ChangeAvatar from "./dialogs/ChangeAvatar.svelte";
    import Textfield from "@smui/textfield";
    import CharacterCounter from "@smui/textfield/character-counter";
    import FormField from "@smui/form-field";
    import Switch from "@smui/switch";
    import ColorPicker from "../../components/ColorPicker.svelte";
    import Button from "@smui/button";

    // Avatar-specific values
    let changeAvatarOpen = false;
    let avatarImageDiv: HTMLDivElement = undefined;

    function reloadAvatarAfterUpload(src: string) {
        avatarImageDiv.style.backgroundImage = `url(${src})`;
    }

    // User data
    let forename = $data.userData.user.forename;
    let surname = $data.userData.user.surname;

    let schedulerEnabled = $data.userData.user.schedulerEnabled;
    let darkTheme = $data.userData.user.darkTheme;

    let primaryColorDark = $data.userData.user.primaryColorDark;
    let primaryColorLight = $data.userData.user.primaryColorLight;
</script>

<ChangeAvatar
    bind:open={changeAvatarOpen}
    on:update={(e) => reloadAvatarAfterUpload(e.detail)}
/>

<div class="preview ">
    <div class="preview__avatar">
        <div class="preview__avatar__image" bind:this={avatarImageDiv} />
        <div class="preview__avatar__edit">
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
        <div class="inputs__danger__delete__user">
            <div>
                <span class="--clr-text-hint"
                    >Erase all user data and delete account</span
                >
            </div>
            <Button variant="outlined">Delete</Button>
        </div>
    </div>
    <div class="inputs__actions">
        <Button>Cancel</Button>
        <Button>Apply Changes</Button>
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
            margin-top: 2rem;
        }

        &__name {
            display: flex;
            gap: 2rem;
        }

        &__toggles {
            background-color: var(--clr-height-1-3);
            border-radius: 0.3rem;
            padding: 1rem;
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
        }

        &__actions {
            display: flex;
            justify-content: flex-end;
            gap: 0.5rem;
            margin-top: 1.5rem;
        }
    }
</style>
