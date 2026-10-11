import { Layout as BasicLayout } from "@rspress/core/theme-original";

const Layout = () => (
    <BasicLayout
        afterNavMenu={
            <div className="catl-nav-actions">
                <a
                    href="/catl/guide/installation"
                    className="catl-button catl-button--primary"
                >
                    Install
                </a>
            </div>
        }
    />
);

export * from "@rspress/core/theme-original";
export { Layout };

