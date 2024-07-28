import { useState, useEffect } from "react";
import PropTypes from 'prop-types';
import "./Referrals.css";
import SentRefs from "./SentRefs";
import ReceivedRefs from "./ReceivedRefs";
import CreateReferral from "./CreateReferral";
import axios from "axios";

const Referrals = ({ role }) => {
  const [activeTab, setActiveTab] = useState("Referrals Sent");
  const [userCompanyId, setUserCompanyId] = useState(null);
  const [loading, setLoading] = useState(true); // Added loading state
  const [error, setError] = useState(""); // Added error state

  useEffect(() => {
    const fetchUserCompanyId = async () => {
      try {
        const response = await axios.get("http://localhost:8080/get-user-info", { withCredentials: true });
        setUserCompanyId(response.data.companyId);
      } catch (error) {
        console.error("Error fetching user info:", error);
        setError("Failed to fetch user information."); // Set error message
      } finally {
        setLoading(false); // Set loading to false after fetching
      }
    };

    fetchUserCompanyId();
  }, []);

  const renderContent = () => {
    if (loading) return <p>Loading...</p>; // Show loading message if loading

    if (error) return <p>{error}</p>; // Show error message if there's an error

    switch (activeTab) {
      case "Referrals Sent":
        return <SentRefs role={role} />;
      case "Referrals Received":
        return <ReceivedRefs role={role} />;
      case "Create Referral":
        return userCompanyId ? (
          <CreateReferral userCompanyId={userCompanyId} />
        ) : (
          <p>Company information not available.</p> // Handle case where userCompanyId is still null
        );
      default:
        return <p>Invalid tab selected.</p>; // Handle unexpected tab values
    }
  };

  return (
    <div className='Referrals'>
      <h2 className='ref-title'>MY REMUNERATION</h2>
      <div className='ref-underline'></div>
      <div className='tab-buttons'>
        <button
          onClick={() => setActiveTab("Referrals Sent")}
          className={activeTab === "Referrals Sent" ? "active" : ""}>
          Referrals Sent
        </button>
        <button
          onClick={() => setActiveTab("Referrals Received")}
          className={activeTab === "Referrals Received" ? "active" : ""}>
          Referrals Received
        </button>
        <button
          onClick={() => setActiveTab("Create Referral")}
          className={activeTab === "Create Referral" ? "active" : ""}>
          Create Referral
        </button>
      </div>
      <div className='tableHeader'>{activeTab}</div>
      <div className='tab-content'>{renderContent()}</div>
      <div className='tab-content-bottom'></div>
    </div>
  );
};

Referrals.propTypes = {
  role: PropTypes.string.isRequired,
};

export default Referrals;
