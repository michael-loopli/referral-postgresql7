import { useEffect, useState } from "react";
import axios from "axios";
import "../styles.css";

const ReceivedRefs = () => {
  const [referrals, setReferrals] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchReceivedReferrals = async () => {
      try {
        const response = await axios.get(
          "http://localhost:8080/referrals-received",
          { withCredentials: true }
        );
        setReferrals(response.data);
      } catch (error) {
        console.error("Error fetching received referrals:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchReceivedReferrals();
  }, []);

  const handleAction = async (referralRequestID, action) => {
    try {
      await axios.post(
        `http://localhost:8080/referral-request-action/${action}/${referralRequestID}`,
        {},
        {
          withCredentials: true,
        }
      );
      alert(`Referral request ${action}ed successfully`);
      setReferrals((prevRequests) =>
        prevRequests.map((request) =>
          request.id === referralRequestID
            ? {
                ...request,
                status: action === "approve" ? "Approved" : "Denied",
              }
            : request
        )
      );
    } catch (error) {
      console.error(`Error ${action}ing referral request:`, error);
      alert(`Failed to ${action} referral request`);
    }
  };

  if (loading) return <p>Loading...</p>;

  return (
    <div>
      <h2>Received Referrals</h2>
      {referrals.length === 0 ? (
        <p>No referral requests sent.</p>
      ) : (
        <ul>
          {referrals.map((referral) => (
            <li key={referral.id}>
              <strong>Referee Client:</strong> {referral.referee_client}
              <br />
              <strong>Title:</strong> {referral.title}
              <br />
              <strong>Context:</strong> {referral.content}
              <br />
              <strong>Referee Client Email:</strong>{" "}
              {referral.referee_client_email}
              <br />
              <strong>Referrer:</strong> {referral.referrer_username} :{" "}
              {referral.company_name}
              <br />
              {/* <strong>Company:</strong> {referral.company_name}
              <br /> */}
              <strong>Status:</strong> {referral.status}
              <br />
              <strong>Created At:</strong>{" "}
              {new Date(referral.created_at).toLocaleString()}
              <br />
              <button
                className='btn-accept'
                onClick={() => handleAction(referral.id, "approve")}>
                Accept
              </button>
              <button
                className='btn-deny'
                onClick={() => handleAction(referral.id, "deny")}>
                Deny
              </button>
              <hr />
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default ReceivedRefs;

{
  /* <strong>{referral.title}</strong>
              <p>{referral.content}</p>
              <p>Referee Client: {referral.referee_client}</p>
              <p>Referee Client Email: {referral.referee_client_email}</p>
              <p>
                Referrer: {referral.referrer_username} : {referral.company_name}
              </p>
              <p>Status: {referral.status}</p>
              <p>
                Created At: {new Date(referral.created_at).toLocaleString()}
              </p> */
}
