// Tabs.js
import  { useState } from "react";

const Tabs = () => {
  const [activeTab, setActiveTab] = useState("Tab1");

  const renderContent = () => {
    switch (activeTab) {
      case "Tab1":
        return <div>Content for Tab 1</div>;
      case "Tab2":
        return <div>Content for Tab 2</div>;
      case "Tab3":
        return <div>Content for Tab 3</div>;
      default:
        return <div>Content for Tab 1</div>;
    }
  };

  return (
    <div className="tabs">
      <div className="tab-buttons">
        <button
          onClick={() => setActiveTab("Tab1")}
          className={activeTab === "Tab1" ? "active" : ""}
        >
          Tab 1
        </button>
        <button
          onClick={() => setActiveTab("Tab2")}
          className={activeTab === "Tab2" ? "active" : ""}
        >
          Tab 2
        </button>
        <button
          onClick={() => setActiveTab("Tab3")}
          className={activeTab === "Tab3" ? "active" : ""}
        >
          Tab 3
        </button>
      </div>
      <div className="tab-content">{renderContent()}</div>
    </div>
  );
};

export default Tabs;
