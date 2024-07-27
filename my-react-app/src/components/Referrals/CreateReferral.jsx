import { useState, useEffect } from 'react';
import axios from 'axios';
import PropTypes from 'prop-types';
import './CreateRefs.css';

const CreateReferral = ({ userCompanyId }) => {
  const [formData, setFormData] = useState({
    title: "",
    content: "",
    sent_company_id: userCompanyId, // Default to user's company_id
    receiving_company_id: "", // Add field for receiving company ID
    referee_client: "",
    referee_client_email: "",
  });
  const [companies, setCompanies] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Fetch the list of companies from the backend
    const fetchCompanies = async () => {
      try {
        const response = await axios.get("http://localhost:8080/companies", { withCredentials: true });
        setCompanies(response.data);
      } catch (error) {
        console.error("Error fetching companies:", error);
      } finally {
        setLoading(false);
      }
    };
    fetchCompanies();
  }, []);

  const handleChange = (e) => {
    setFormData({
      ...formData,
      [e.target.name]: e.target.name === "sent_company_id" || e.target.name === "receiving_company_id"
        ? parseInt(e.target.value, 10)
        : e.target.value,
    });
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    console.log('Form Data:', formData); // Log form data to inspect the values
    try {
      const response = await axios.post(
        "http://localhost:8080/create-referral",
        formData,
        {
          headers: {
            "Content-Type": "application/json",
          },
          withCredentials: true,
        }
      );
      if (response.status === 201) {
        alert("Referral request created successfully");
        setFormData({
          title: "",
          content: "",
          sent_company_id: userCompanyId, // Reset to user's company_id
          receiving_company_id: "", // Reset to empty
          referee_client: "",
          referee_client_email: "",
        });
      }
    } catch (error) {
      console.error("Error creating referral request:", error);
      alert("Failed to create referral request: " + (error.response?.data || error.message));
    }
  };
  
  if (loading) return <p>Loading companies...</p>;

  return (
    <div className='create-referral'>
      <h1>Create Referral</h1>
      <form className='referral-form' onSubmit={handleSubmit}>
        <div className='refs'>
          <div>
            <div className='ref-input'>
              <label htmlFor='title'>Nature of Referral:</label>
              <input
                type='text'
                id='title'
                name='title'
                value={formData.title}
                onChange={handleChange}
                required
              />
            </div>
          </div>

          <br />

          <label htmlFor='content'>Context of Referral:</label>
          <textarea
            id='content'
            name='content'
            value={formData.content}
            onChange={handleChange}
            required
          />

          <br />

          <div className='ref-dropdown'>
            <label htmlFor='sentCompany'>Sending Company:</label>
            <select
              id='sentCompany'
              name='sent_company_id'
              value={formData.sent_company_id}
              onChange={handleChange}
              required
              disabled
            >
              <option value={userCompanyId}>{`Company ID: ${userCompanyId}`}</option> {/* Display user's company ID */}
            </select>
          </div>

          <br />

          <div className='ref-dropdown'>
            <label htmlFor='receivingCompany'>Receiving Company:</label>
            <select
              id='receivingCompany'
              name='receiving_company_id'
              value={formData.receiving_company_id}
              onChange={handleChange}
              required
            >
              <option value=''>Select a company</option>
              {companies.map((company) => (
                <option key={company.id} value={company.id}>
                  {company.name}
                </option>
              ))}
            </select>
          </div>

          <br />

          <div className='ref-input'>
            <label htmlFor='refereeClient'>Referee Client Name:</label>
            <input
              type='text'
              id='refereeClient'
              name='referee_client'
              value={formData.referee_client}
              onChange={handleChange}
              required
            />
          </div>
          <br />

          <div className='ref-input'>
            <label htmlFor='refereeClientEmail'>Referee Client Email:</label>
            <input
              type='email'
              id='refereeClientEmail'
              name='referee_client_email'
              value={formData.referee_client_email}
              onChange={handleChange}
              required
            />
          </div>
          <br />
        </div>
        <div className='refsubmit-container'>
          <button className='ref-submit' type='submit'>
            Send Referral
          </button>
        </div>
      </form>
    </div>
  );
};

CreateReferral.propTypes = {
  userCompanyId: PropTypes.number.isRequired,
};

export default CreateReferral;
