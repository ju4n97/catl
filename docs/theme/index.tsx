import { Layout as BasicLayout } from "@rspress/core/theme-original";

const Layout = () => (
    <BasicLayout
        afterNavMenu={
            <div className="regi-nav-actions">
                <a
                    href="/regi/guide/installation"
                    className="regi-button regi-button--primary"
                >
                    Install
                </a>
            </div>
        }
    />
);

export * from "@rspress/core/theme-original";
export { Layout };

