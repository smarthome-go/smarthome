<script lang="ts">
    import { data, fetchData } from "../global";
    import NavBarButton from "./NavBarButton.svelte";
    import NotificationDrawer from "./NotificationDrawer.svelte";
    import { get } from "svelte/store";
    import { onMount } from 'svelte'

    export let persistentClose: boolean = true
    export let closed = true;

    const toggleClosed = () => {
        closed = !closed
    };

    let drawerClosed = true;

    let nav: HTMLElement | null;
    document.addEventListener(
        "click",
        (event) => {
            if (nav === null || !nav.contains(event.target as Node)) {
                closed = true;
                drawerClosed = true;
            }
        },
        true
    );

    let avatar: HTMLDivElement;

    interface Page {
        label: string;
        uri: string;
        icon: string;
        position: "top" | "bottom";
        permission: string; // '' means no permission required
    }

    let pagesFiltered: Page[] = [];
    const pages: Page[] = [
        {
            label: "Dashboard",
            uri: "/dash",
            icon: "home",
            position: "top",
            permission: "",
        },
        {
            label: "Rooms",
            uri: "/rooms",
            icon: "view_quilt",
            position: "top",
            permission: "setPower",
        },
        {
            label: "Reminders",
            uri: "/reminders",
            icon: "task_alt",
            position: "top",
            permission: "reminder",
        },
        {
            label: "Scheduler",
            uri: "/scheduler",
            icon: "schedule",
            position: "top",
            permission: "scheduler",
        },
        {
            label: "Automation",
            uri: "/automations",
            icon: "event_repeat",
            position: "top",
            permission: "automation",
        },
        {
            label: "Homescript",
            uri: "/homescript",
            icon: "terminal",
            position: "top",
            permission: "homescript",
        },
        {
            label: "Profile",
            uri: "/profile",
            icon: "manage_accounts",
            position: "top",
            permission: "",
        },
        {
            label: "Users",
            uri: "/users",
            icon: "admin_panel_settings",
            position: "bottom",
            permission: "manageUsers",
        },
        {
            label: "System",
            uri: "/system",
            icon: "settings",
            position: "bottom",
            permission: "modifyServerConfig",
        },
        {
            label: "Logout",
            uri: "/logout",
            icon: "logout",
            position: "bottom",
            permission: "",
        },
    ];

    function withoutPosition(page: Page): {
        label: string;
        uri: string;
        icon: string;
    } {
        return {
            label: page.label,
            uri: page.uri,
            icon: page.icon,
        };
    }

    onMount(async () => {
        await fetchData();
        // Filter out any pages to which the user has no access to
        pagesFiltered = pages.filter(
            (p) =>
                $data.userData.permissions.includes(p.permission) ||
                p.permission == "" ||
                $data.userData.permissions.includes("*")
        );

        avatar.style.backgroundImage = `url(/api/user/avatar/personal?urscache=${get(data).userData.user.username}`
    });
</script>

<nav bind:this={nav} class:closed class:persistentClose={persistentClose && closed}>
    <div
        id="bg"
        class:mdc-elevation--z16={drawerClosed}
        class:mdc-elevation--z8={!drawerClosed}
    />
    <div id="toggle" class:visible={persistentClose} on:click={toggleClosed} on:keydown={toggleClosed}>
        <i class="material-icons">chevron_right</i>
    </div>
    <div id="header">
        <div bind:this={avatar} id="header__avatar"/>
        <div id="header__texts">
            <strong
                >{$data.userData.user.forename}
                {$data.userData.user.surname}</strong
            >
            <span>{$data.userData.user.username}</span>
        </div>
    </div>
    <div id="bell" on:click={() => (drawerClosed = !drawerClosed)} on:keydown={() => (drawerClosed = !drawerClosed)}>
        <div id="bell__icon">
            <div id="bell__icon__inner">
                <i class="material-icons"
                    >{$data.notificationCount === 0
                        ? "notifications"
                        : "notifications_active"}</i
                >
                <div class:hidden={$data.notificationCount === 0}>
                    <span>{$data.notificationCount}</span>
                </div>
            </div>
        </div>
        <span id="bell__text"
            >{"Notification" + ($data.notificationCount !== 1 ? "s" : "")}</span
        >
    </div>
    <NotificationDrawer bind:hidden={drawerClosed} />
    <div id="menubar">
        <div id="menubar__top">
            {#each pagesFiltered.filter((p) => p.position === "top") as page}
                <NavBarButton
                    {...withoutPosition(page)}
                    active={page.uri === window.location.pathname}
                />
            {/each}
        </div>
        <div id="menubar__bottom">
            {#each pagesFiltered.filter((p) => p.position === "bottom") as page}
                <NavBarButton
                    {...withoutPosition(page)}
                    active={page.uri === window.location.pathname}
                />
            {/each}
        </div>
    </div>
</nav>

<style lang="scss">
    @use "../mixins" as *;

    nav {
        position: fixed;
        left: 0;
        top: 0;
        bottom: 0;
        width: 13rem;
        user-select: none;
        box-sizing: border-box;
        padding: 1rem;
        display: flex;
        flex-direction: column;
        & > * {
            flex-shrink: 0;
        }
        white-space: nowrap;
        transition-property: width, height;
        transition-duration: 0.3s;
        z-index: 100;

        // Hide visible overflow when closed
        &.closed, &.persistentClose {
            #menubar {
                @include mobile {
                    overflow: hidden;
                }
            }
        }

        &.persistentClose {
            width: 5.125rem;

            @include mobile {
                width: auto;
                height: 3.5rem;
            }
        }

        @include mobile {
            bottom: auto;
            width: auto;
            right: 0;
            padding-top: 0;
            height: 100%;
        }

        @include not-widescreen {
            &.closed {
                width: 5.125rem;

                @include mobile {
                    width: auto;
                    height: 3.5rem;
                }
            }
        }
    }

    #bg {
        position: absolute;
        inset: 0;
        transition-property: box-shadow;
        transition-duration: 0.3s;
        background-color: var(--clr-height-0-16);
    }

    #toggle {
        position: absolute;
        right: 0;
        top: 50%;
        color: var(--clr-text);
        background-color: var(--clr-hover);
        border-radius: 50%;
        cursor: pointer;
        height: 2rem;
        aspect-ratio: 1;
        overflow: hidden;
        transform: translate(50%, -50%);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 10;
        transition: opacity 0.3s;

        @include mobile {
            top: 1.75rem;
            left: 5rem;
            transform: translateY(-50%) rotate(90deg);
        }

        &:not(.visible) {
            @include widescreen {
                opacity: 0;
                pointer-events: none;
            }
        }

        i {
            font-size: 1.5rem;
            transform: rotate(180deg);
            transition: transform 0.3s;

            nav.closed & {
                transform: rotate(0deg);
            }
        }
    }

    #header {
        position: relative;
        display: flex;
        align-items: center;
        gap: 0.6rem;
        padding-block: 0.5rem;
        padding-left: 0.4rem;
        overflow-x: hidden;
        min-height: 3.5rem;

        @include mobile {
            width: min-content;
            min-height: auto;
        }

        &__avatar {
            background-position: center;
            background-size: cover;
            background-repeat: no-repeat;
            border-radius: 50%;
            aspect-ratio: 1;
            height: 2.5rem;
            // background-image: url("/api/user/avatar/personal");
        }
        &__texts {
            display: flex;
            flex-direction: column;
            gap: 0.2rem;
            pointer-events: none;

            @include mobile {
                display: none;
            }

            &:first-child {
                font-weight: 600;
            }

            strong,
            span {
                overflow: hidden;
                white-space: nowrap;
                text-overflow: ellipsis;
                max-width: 7rem;
            }
        }
    }

    #bell {
        position: relative;
        overflow-x: hidden;
        border-radius: 0.4rem;
        display: flex;
        align-items: center;
        gap: 0.3rem;
        height: 3.125rem;
        cursor: pointer;
        transition: background-color 0.2s;

        &:hover {
            background-color: var(--clr-hover);
        }
        @include mobile {
            width: min-content;
            position: absolute;
            top: 1.75rem;
            right: 1rem;
            transform: translate(200%, -50%);
            transition: transform 0.2s;

            nav.closed & {
                transform: translateY(-50%);
            }
        }

        &__icon {
            height: 100%;
            aspect-ratio: 1;
            display: flex;
            align-items: center;
            justify-content: center;

            &__inner {
                line-height: 0.75;
                position: relative;
                i {
                    font-size: var(--icon-size);
                }
                div {
                    position: absolute;
                    font-size: 0.6rem;
                    border-radius: 50%;
                    background-color: var(--clr-primary);
                    height: 0.8rem;
                    aspect-ratio: 1;
                    padding: 0.1rem;
                    top: 0;
                    right: 0;
                    transform: translate(50%, -50%);
                    display: flex;
                    align-items: center;
                    justify-content: center;
                    transition-property: opacity;
                    transition-duration: 0.2s;

                    &.hidden {
                        opacity: 0;
                    }
                    span {
                        color: var(--clr-on-primary);
                    }
                }
            }
        }

        &__text {
            @include mobile {
                display: none;
            }
        }
    }

    #menubar {
        position: relative;
        display: flex;
        justify-content: space-between;
        flex-direction: column;
        overflow-x: hidden;
        flex-grow: 1;
        padding-top: 1rem;

        @include mobile {
            flex-shrink: 1;
            //overflow-y: hidden;

            &__top {
                @include landscape {
                    display: flex;
                    flex-wrap: wrap;
                    flex-direction: column;
                    height: 50vh;
                }
            }

            &__bottom {
                @include landscape {
                    display: flex;
                    gap: 1rem;
                    background-color: var(--clr-hover);
                    border-radius: 0.3rem;
                    padding-bottom: 0.625rem;
                }
            }
        }
    }
</style>
