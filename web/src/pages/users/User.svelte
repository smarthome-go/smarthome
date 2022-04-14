<script lang="ts">
    import Button,{ Icon,Label } from '@smui/button'

    export let username: string
    export let forename: string
    export let surname: string
    export let primaryColorDark: string
    export let primaryColorLight: string
    export let schedulerEnabled: boolean
    export let darkTheme: boolean
</script>

<div id="root">
    <div id="left">
        <img
            src={`/api/user/avatar/user/${username}`}
            alt="the users avatar"
            class="mdc-elevation--z4"
        />
        <div id="labels">
            <h6>{username}</h6>
            <span>{forename} {surname}</span>
        </div>
    </div>

    <div id="manage">
        <Button on:click={() => alert(`popup should open instead`)}>
            <Label>Manage</Label>
            <Icon class="material-icons">edit</Icon>
        </Button>
    </div>
</div>

<style lang="scss">
    @use '../../mixins' as *;
    #root {
        background-color: var(--clr-height-1-3);
        border-radius: 0.3rem;
        padding: 1rem;
        display: flex;
        align-items: center;
        justify-content: space-between;
        @include mobile {
            flex-direction: column;
            align-items: normal;
        }
    }
    #left {
        display: flex;
        gap: 1rem;
    }
    #manage {
        @include mobile {
            display: block;
        }
    }
    #labels {
        h6 {
            display: block;
            margin: 0;
        }
        span {
            @include mobile {
                display: none;
            }
        }
    }
    img {
        border-radius: 50%;
        width: 4rem;
        height: 4rem;

        @include mobile {
            width: 2rem;
            height: 2rem;
        }
    }
</style>
